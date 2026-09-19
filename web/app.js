"use strict";

const form = document.querySelector("#parse-form");
const filenameInput = document.querySelector("#filename");
const status = document.querySelector("#status");
const result = document.querySelector("#result");
const outputFields = document.querySelectorAll("[data-field]");

function setStatus(message, error = false) {
  status.textContent = message;
  status.classList.toggle("error", error);
  status.hidden = !message;
}

function setUnavailable(message) {
  filenameInput.disabled = true;
  setStatus(message, true);
}

function render(info) {
  for (const output of outputFields) {
    const value = info[output.dataset.field];
    const empty = value === "" || value === 0 || value == null;
    output.textContent = empty ? "—" : String(value);
  }
  result.hidden = false;
}

async function instantiate(go) {
  const response = await fetch("./whatsthis.wasm");
  if (!response.ok) {
    throw new Error(`could not load WebAssembly (${response.status})`);
  }

  if (WebAssembly.instantiateStreaming) {
    try {
      return await WebAssembly.instantiateStreaming(response.clone(), go.importObject);
    } catch (error) {
      console.warn("Streaming WebAssembly compilation failed; falling back to ArrayBuffer.", error);
    }
  }

  return WebAssembly.instantiate(await response.arrayBuffer(), go.importObject);
}

form.addEventListener("submit", (event) => {
  event.preventDefault();

  const filename = filenameInput.value.trim();
  if (!filename) {
    return;
  }

  try {
    const encoded = globalThis.whatsthis.video(filename);
    if (!encoded) {
      throw new Error("the parser returned no result");
    }
    render(JSON.parse(encoded));
    setStatus("");
  } catch (error) {
    console.error(error);
    setStatus("Could not parse filename.", true);
  }
});

(async () => {
  if (globalThis.location?.protocol === "file:") {
    setUnavailable("Serve dist over HTTP to load WebAssembly.");
    return;
  }

  if (!globalThis.WebAssembly || !globalThis.Go) {
    setUnavailable("WebAssembly is not supported.");
    return;
  }

  try {
    const go = new Go();
    const { instance } = await instantiate(go);
    void go.run(instance).then(() => {
      setUnavailable("WebAssembly stopped unexpectedly.");
    }).catch((error) => {
      console.error(error);
      setUnavailable("WebAssembly stopped unexpectedly.");
    });

    if (typeof globalThis.whatsthis?.video !== "function") {
      throw new Error("the parser API was not registered");
    }

    filenameInput.disabled = false;
    setStatus("");
    filenameInput.focus();
  } catch (error) {
    console.error(error);
    setUnavailable("Could not initialize WebAssembly.");
  }
})();

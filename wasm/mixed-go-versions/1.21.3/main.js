import wasmExec from "./wasm_exec_noglobals.js";



async function init() {
	const go = new wasmExec.Go();
	let result = await WebAssembly.instantiateStreaming(fetch("1.21.3/main.wasm"), go.importObject)
	go.run(result.instance);
}
init();

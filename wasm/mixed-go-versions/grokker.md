in wasm_exec.js, where does the 'globalThis.TextEncoder' come from?

The 'globalThis.TextEncoder' in wasm_exec.js does not actually come
from wasm_exec.js itself. Instead, it is a part of the global
JavaScript runtime environment in which the JavaScript code is being
run.

The 'TextEncoder' class is an API provided by the browser's JavaScript
environment that's used for encoding a string of any character
encoding scheme into UTF-8. It's a part of the Encoding API, and
globalThis refers to the global object, whether that's the window, the
global, or something else.

Therefore, when wasm_exec.js is being used, it expects that the
'globalThis.TextEncoder' already exists in the environment as a global
object. If it doesn't exist, an error will be thrown, indicating that
a polyfill is required.


Uncaught SyntaxError: Unexpected reserved word (at index.js:6:15)
why would 'await' be a reserved word?

The `await` keyword is a reserved word in JavaScript because it's used
with the `async`/`await` feature which is part of ES2017 syntax to
work with promises in a more comfortable synchronous way. If you use
`await` in non-async functions or outside of a function in the
top-level code, you will get a syntax error because JavaScript isn't
expecting `await` to be used there.


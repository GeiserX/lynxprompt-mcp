#!/usr/bin/env node
"use strict";

const { spawn } = require("child_process");
const { install } = require("./postinstall.js");

// npm 12 no longer runs postinstall by default, so the binary may not be
// there yet: fetch and verify it on first use.
install()
  .then((binPath) => {
    const env = Object.assign({}, process.env, { TRANSPORT: "stdio" });
    const child = spawn(binPath, process.argv.slice(2), { env, stdio: "inherit" });
    for (const sig of ["SIGINT", "SIGTERM"]) {
      process.on(sig, () => child.kill(sig));
    }
    child.on("exit", (code, signal) => {
      process.exit(signal ? 1 : code ?? 1);
    });
    child.on("error", (err) => {
      process.stderr.write(`lynxprompt-mcp: could not start ${binPath}: ${err.message}\n`);
      process.exit(1);
    });
  })
  .catch((err) => {
    process.stderr.write(`lynxprompt-mcp: ${err.message}\n`);
    process.exit(1);
  });

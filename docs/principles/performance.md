# Performance principles

1. **Hot path is local binary IPC.** No HTTP or JSON on keystroke handling.
2. **Bound work per key.** Dictionary load and heavy index builds are offline or startup; ranking stays budgeted.
3. **Measure before claiming.** Latency and throughput numbers belong under `docs/validation/performance.md` with method.
4. **Avoid cross-process chatter.** Batch only when it reduces round-trips without harming responsiveness.
5. **Adapter stays thin.** UI and host callbacks must not block on engine-scale work inside the adapter process.

# Runtime diagnostics

Run the browser UI with:

```bash
devin-web -log=on
```

When diagnostics are enabled, the browser records the theme-flow stages in
`logs/theme-flow.jsonl`: polygon ready, export, download, main UI initialization,
file selection, theme application, and persistence.

The file is intentionally stored inside the project so a UX test run is easy to
find and inspect. JSONL files are ignored by Git. Logs contain only timestamps,
event names, theme name/version, and selected file metadata; prompts, workspace
contents, credentials, and cookies are not written.

To inspect the flow:

```bash
cat logs/theme-flow.jsonl
```

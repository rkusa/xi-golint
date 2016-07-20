# xi-golint

```bash
go build -o ~/Development/forks/xi-editor/xigolint ./*.go
```

in `run_plugin.rs`

```diff
-       pathbuf.push("python");
-       pathbuf.push("plugin.py");
+       pathbuf.push("xigolint");
```

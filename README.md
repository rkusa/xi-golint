# xi-golint

```bash
git clone https://github.com/rkusa/xi-editor
git clone https://github.com/rkusa/xi-golint

cd xi-golint
go build -o ../xi-editor/xi-golint ./*.go


cd ../xi-editor
git checkout xi-golint
xcodebuild
open build/Release/XiEditor.app
```

Hit F3 to run the plugin

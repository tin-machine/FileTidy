# FileTidy

# インストール

```bash
go install github.com/tin-machine/filetidy@latest
```

# 使い方

コマンドを実行するカレントディレクトリにあるconfig.ymlを読み込み実行

```bash
filetidy
```

# 設定

sourcepath: 整理対象のディレクトリを指定
destinationpath: 整理後のディレクトリを指定
extentiontarget: 拡張子ごとのフォルダ名を指定
filenameregex: ファイル名にマッチする正規表現を指定

```yaml
sourcepath:
  - ~/Downloads/
  - ~/Desktop/

destinationpath: ~/workspace/Archive

filenameregex:
  スクリーンショット\ (\d{4}-\d{2}-\d{2}) .*png: ScreenShot

extentiontarget:
  txt: TXT
  jpeg: Image
  JPEG: Image
  pdf: PDF
  PDF: PDF
```

# Options

デバックオプション、実行結果の確認

```bash
filetidy -d
```

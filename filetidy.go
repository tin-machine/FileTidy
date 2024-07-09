package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/k0kubun/pp"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"gopkg.in/go-playground/validator.v9"
)

type Config struct {
	Sourcepath      []string          `validate:"required"`
	Destinationpath string            `validate:"required"`
	Extentiontarget map[string]string `validate:"required"`
	Filenameregex   map[string]string `validate:"required"`
}

// expandPathはチルダをホームディレクトリに展開し、相対パスを絶対パスに変換する
func expandPath(path string) (string, error) {
	// チルダをホームディレクトリに展開する
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(homeDir, path[1:])
	}

	// 相対パスを絶対パスに変換する
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return absPath, nil
}

func listFiles(dirs []string) []string {
	var fileList []string
	for _, dir := range dirs {
		// パスを展開する
		expandedDir, err := expandPath(dir)
		if err != nil {
			panic(err)
		}

		files, err := os.ReadDir(expandedDir)
		if err != nil {
			panic(err)
		}
		for _, file := range files {
			if !file.IsDir() {
				fileList = append(fileList, filepath.Join(expandedDir, file.Name()))
			}
		}
	}
	return fileList
}

func Sjis2Utf8(str string) (string, error) {
	iostr := strings.NewReader(str)
	rio := transform.NewReader(iostr, japanese.ShiftJIS.NewDecoder())
	ret, err := io.ReadAll(rio)
	if err != nil {
		return "", err
	}
	return string(ret), err
}

func Utf82Sjis(str string) (string, error) {
	iostr := strings.NewReader(str)
	rio := transform.NewReader(iostr, japanese.ShiftJIS.NewEncoder())
	ret, err := io.ReadAll(rio)
	if err != nil {
		return "", err
	}
	return string(ret), err
}

func createDir(newPath string) {
	targetDir := filepath.Dir(newPath)
	if f, err := os.Stat(targetDir); os.IsNotExist(err) || !f.IsDir() {
		fmt.Println("移動先のディレクトリはありません。 " + targetDir + " 作ります。")
		if err := os.MkdirAll(targetDir, 0777); err != nil {
			fmt.Println(err)
		}
	}
}

func mv(oldPath string, newPath string) {
	fmt.Println("mv です " + oldPath + " を " + newPath + " に移動します")
	r, err := os.Open(oldPath)
	if err != nil {
		log.Fatal(err)
	}
	w, err := os.Create(newPath)
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(w, r)
	if err != nil {
		log.Fatal(err)
	}
	fileinfo, _ := os.Stat(oldPath)
	err = os.Chmod(newPath, fileinfo.Mode())
	if err != nil {
		log.Fatal(err)
	}
	err = os.Remove(oldPath)
	if err != nil {
		log.Fatal(err)
	}
}

func fileMove(files []string, config Config) {
	// Destinationpathを展開する
	destinationPath, err := expandPath(config.Destinationpath)
	if err != nil {
		log.Fatalf("failed to expand destination path: %v", err)
	}

	// files の各ファイルについて処理を行う
	for _, file := range files {
		fmt.Println("fileMove2 file は " + file)

		// config.Filenameregex 内の各正規表現とファイル名を照合する
		for key, value := range config.Filenameregex {
			r := regexp.MustCompile(key)
			if s := r.FindStringSubmatch(file); len(s) > 0 {
				// 正規表現にマッチした場合の処理
				fmt.Println("正規表現にマッチする " + key + " " + value + " マッチした文字列 " + s[1])
				newPath := filepath.Join(destinationPath, value, s[1], filepath.Base(file))
				fmt.Println("移行先は " + newPath)
				// 必要なディレクトリを作成
				createDir(newPath)
				// ファイルを新しい場所に移動
				mv(file, newPath)
				// 次のファイルに移動
				continue
			}
		}

		// ファイルの拡張子を確認し、移行先ディレクトリを設定
		if val, ok := config.Extentiontarget[strings.Trim(filepath.Ext(file), ".")]; ok {
			newPath := filepath.Join(destinationPath, val, filepath.Base(file))
			// 必要なディレクトリを作成
			createDir(newPath)
			// ファイルを新しい場所に移動
			mv(file, newPath)
		}
	}
}

func main() {
	// "config.yml"というファイルを読み込む
	config, err := os.ReadFile("config.yml")
	// 読み込み中にエラーが発生した場合、プログラムを停止しエラーを表示
	if err != nil {
		panic(err)
	}
	// バリデータを初期化する
	validate := validator.New()
	// Config構造体の変数を宣言
	var c Config
	// 読み込んだYAMLデータをConfig構造体にデコードする
	err = yaml.Unmarshal(config, &c)
	// デコード中にエラーが発生した場合、ログを出力してプログラムを終了
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	// 構造体のバリデーションを実行する
	err = validate.Struct(c)
	// バリデーションに失敗した場合、ログを出力してプログラムを終了
	if err != nil {
		log.Fatalf("validation error: %v", err)
	}
	// Config構造体の内容を整形して表示する
	pp.Print(c)
	// ソースパスのファイルリストを取得し、ファイル移動処理を実行する
	fileMove(listFiles(c.Sourcepath), c)
}

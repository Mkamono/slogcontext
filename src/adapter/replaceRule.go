package adapter

import (
	"log/slog"
	"runtime"
)

const (
	FunctionKey = "function"
	FileKey     = "file"
	LineKey     = "line"
)

// SourceHierarchyRule ログが呼び出されたファイル名、行番号、関数名をログに追加する
// hierarchy: 呼び出し元の階層数
// デフォルトはslog.InfoContextが呼び出された階層
func SourceHierarchyRule(hierarchy int) ReplaceRule {
	defaultHierarchy := 9
	return func(a slog.Attr) slog.Attr {
		switch a.Key {
		case slog.SourceKey:
			file, line, fn := getSource(defaultHierarchy + hierarchy)
			return slog.Group(slog.SourceKey,
				slog.Attr{Key: FunctionKey, Value: slog.StringValue(fn)},
				slog.Attr{Key: FileKey, Value: slog.StringValue(file)},
				slog.Attr{Key: LineKey, Value: slog.IntValue(line)},
			)
		}
		return a
	}
}

// getSource ログが呼び出されたファイル名、行番号、関数名を取得する
func getSource(hierarchy int) (file string, line int, fn string) {
	pc, file, line, _ := runtime.Caller(hierarchy)
	fn = runtime.FuncForPC(pc).Name()
	return file, line, fn
}

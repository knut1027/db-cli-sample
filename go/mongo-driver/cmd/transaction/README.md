# How to run

```cgo
go run cmd/transaction/main.go -deleted
```

flag

| flag | description |
| --- | --- |
| -deleted | delete all document |


# 検証1
セッションの開始と終了にログを仕込むことで、処理がいつ開始/終了したのかを検証する。

## 同じドキュメントを更新しようとした場合

### 更新対象

| _id | description |
|-----| --- |
| 1   | `{"title": "The Bluest Eye", "author": "Toni Morrison"}` |
| 2   | `{"title": "Sula", "author": "Toni Morrison"}` |
| 3   | `{"title": "Song of Solomon", "author": "Toni Morrison"}` | 

### 実行
ターミナル1
```cgo
$ go run cmd/transaction/main.go -deleted true
{"level":"info","ts":1691025499.3737411,"caller":"transaction/main.go:37","msg":"deleted"}
{"level":"info","ts":1691025499.373814,"caller":"transaction/main.go:102","msg":"start session"}
{"level":"info","ts":1691025499.373833,"caller":"transaction/main.go:83","msg":"insert many..."}
{"level":"info","ts":1691025505.379179,"caller":"transaction/main.go:104","msg":"end session"}
```

ターミナル2
```cgo
$ go run cmd/transaction/main.go
{"level":"info","ts":1691025500.849925,"caller":"transaction/main.go:102","msg":"start session"}
{"level":"info","ts":1691025500.850018,"caller":"transaction/main.go:83","msg":"insert many..."}
{"level":"info","ts":1691025500.8577821,"caller":"transaction/main.go:83","msg":"insert many..."}
...
{"level":"info","ts":1691025505.198784,"caller":"transaction/main.go:83","msg":"insert many..."}
{"level":"info","ts":1691025505.383748,"caller":"transaction/main.go:104","msg":"end session"}
{"level":"error","ts":1691025505.383778,"caller":"transaction/main.go:71","msg":"failed to insert","error":"failed to transact: write exception: write errors: [E11000 duplicate key error collection: test.bookInfo index: _id_ dup key: { _id: \"1\" }]","stacktrace":"main.main\n\t/db-cli-sample/go/mongo-driver/cmd/transaction/main.go:71\nruntime.main\n\t/usr/local/go/src/runtime/proc.go:250"}
```

### 結果
先に実行したセッション（ターミナル1）が終了するまで、他の操作は実行されなかった。


## 違うドキュメントを更新しようとした場合

### 更新対象 

| _id | description |
|-----| --- |
| 自動生成 | `{"title": "The Bluest Eye", "author": "Toni Morrison"}` |
| 自動生成 | `{"title": "Sula", "author": "Toni Morrison"}` |
| 自動生成 | `{"title": "Song of Solomon", "author": "Toni Morrison"}` | 

IDをコメントアウトする。
```
book1 := Book{
// 	ID:     "1",
	Title:  "The Bluest Eye",
	Author: "Toni Morrison",
}
```

### 実行
ターミナル1
```cgo
$ go run cmd/transaction/main.go -deleted true
{"level":"info","ts":1691025629.693434,"caller":"transaction/main.go:37","msg":"deleted"}
{"level":"info","ts":1691025629.6935081,"caller":"transaction/main.go:102","msg":"start session"}
{"level":"info","ts":1691025629.693525,"caller":"transaction/main.go:83","msg":"insert many..."}
{"level":"info","ts":1691025635.7043269,"caller":"transaction/main.go:104","msg":"end session"}
```

ターミナル2
```cgo
$ go run cmd/transaction/main.go
{"level":"info","ts":1691025630.77773,"caller":"transaction/main.go:102","msg":"start session"}
{"level":"info","ts":1691025630.777865,"caller":"transaction/main.go:83","msg":"insert many..."}
{"level":"info","ts":1691025636.796876,"caller":"transaction/main.go:104","msg":"end session"}
```


### 結果
先に実行したセッション（ターミナル1）と同時に処理が実行された。

package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"math/rand"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"log"
)

// Food 構造体の定義
type Food struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	// データベースファイルのパス
	dbPath := "./example.db"

	// データベースに接続
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// foods テーブルが存在しない場合は作成
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS foods (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL
        )`)
	if err != nil {
		log.Fatal(err)
	}

	// 乱数生成器の初期化
	rand.Seed(time.Now().UnixNano())

	// ルートパスで静的ファイルを提供
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// /query で foods テーブルのデータを取得して JSON 形式で返すエンドポイント
	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		// foods テーブルからデータを取得
		rows, err := db.Query("SELECT id, name FROM foods")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		// 取得したデータを Food 構造体にマッピングして foodsInserted に追加
		var foodsInserted []Food
		for rows.Next() {
			var f Food
			err = rows.Scan(&f.ID, &f.Name)
			if err != nil {
				log.Fatal(err)
			}
			foodsInserted = append(foodsInserted, f)
		}

		// JSON形式でレスポンスを返す
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(foodsInserted)
	})

	// /insert で POST リクエストを受け取り、foods テーブルにデータを挿入するエンドポイント
	http.HandleFunc("/insert", func(w http.ResponseWriter, r *http.Request) {
		// POSTリクエストからデータを取得
		name := r.FormValue("name")

		// foods テーブルにデータを挿入
		_, err := db.Exec("INSERT INTO foods (name) VALUES (?)", name)
		if err != nil {
			log.Fatal(err)
		}

		// 挿入後のデータを再度クエリして JSON 形式でレスポンスを返す
		rows, err := db.Query("SELECT id, name FROM foods")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		// 取得したデータを Food 構造体にマッピングして foodsInserted に追加
		var foodsInserted []Food
		for rows.Next() {
			var f Food
			err = rows.Scan(&f.ID, &f.Name)
			if err != nil {
				log.Fatal(err)
			}
			foodsInserted = append(foodsInserted, f)
		}

		// JSON形式でレスポンスを返す
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(foodsInserted)
	})

	// /reset で foods テーブルの全データを削除する
	http.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		// foods テーブルのデータを削除
		_, err := db.Exec("DELETE FROM foods")
		if err != nil {
			log.Fatal(err)
		}

		// 空の JSON 配列を返す
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]")) // 空の JSON 配列を返す
	})

	// /roulette でランダムに name を取得するエンドポイント
	http.HandleFunc("/roulette", func(w http.ResponseWriter, r *http.Request) {
		// foods テーブルからランダムに1つの name を取得する
		var name string
		err := db.QueryRow("SELECT name FROM foods ORDER BY RANDOM() LIMIT 1").Scan(&name)
		if err != nil {
			log.Fatal(err)
		}

		// JSON形式でレスポンスを返す
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"name": name})
	})

	// HTTPサーバーを起動
	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

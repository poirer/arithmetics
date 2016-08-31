// This is an attempt to play with database on a rather low level. Since it is not a topic for now, you may skip it while reviewing
package main

import (
	"database/sql"
	"errors"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func connect(dbUrl string) *sql.DB {
	db, err := sql.Open("sqlite3", dbUrl)
	if err != nil {
		log.Fatal("Cannot connect to database", err)
	}
	return db
}

func findWordId(db *sql.DB, user, word string) (int, error) {
	var id int
	var row *sql.Row = db.QueryRow("select id from Words where word = ? and owner = ?", word, user)
	err := row.Scan(&id)
	return id, err
}

func addDictEntry(db *sql.DB, user string, dictEntry dictionaryEntry) error {
	// It might make sense to manage transaction here. But it is too advanced for now
	_, err := findWordId(db, user, dictEntry.Word)
	if err == nil {
		return errors.New("Word already exists")
	}
	res, err := db.Exec("insert into Words(word, owner) values(?, ?)", dictEntry.Word, user)
	if err != nil {
		log.Println("Cannot create new entry", err)
		return err
	}
	pk, err := res.LastInsertId()
	for _, t := range dictEntry.Translations {
		_, err := db.Exec("insert into Translations(translation, word_id) values(?, ?)", t, pk)
		if err != nil {
			log.Println("Error occurred while saving translations", err)
			return err
		}
	}
	for _, i := range dictEntry.Idioms {
		_, err := db.Exec("insert into Idioms(idiom, word_id) values(?, ?)", i, pk)
		if err != nil {
			log.Println("Error occurred while saving idioms", err)
			return err
		}
	}
	return nil
}

func updateDictEntry(db *sql.DB, user string, dictEntry dictionaryEntry) error {
	word_id, err := findWordId(db, user, dictEntry.Word)
	if err != nil {
		log.Println("Cannot obtain word to update", err)
		return err
	}
	_, err = db.Exec("delete from Translations where word_id = ?", word_id)
	if err != nil {
		log.Println("Error occurred while updating dictionary entry", err)
		return err
	}
	for _, t := range dictEntry.Translations {
		_, err := db.Exec("insert into Translations(translation, word_id) values(?, ?)", t, word_id)
		if err != nil {
			log.Println("Error occurred while saving translations", err)
			return err
		}
	}
	_, err = db.Exec("delete from Idioms where word_id = ?", word_id)
	if err != nil {
		log.Println("Error occurred while deleting idioms", err)
		return err
	}
	for _, i := range dictEntry.Idioms {
		_, err := db.Exec("insert into Idioms(idiom, word_id) values(?, ?)", i, word_id)
		if err != nil {
			log.Println("Error occurred while saving idioms", err)
			return err
		}
	}
	return nil
}

func deleteDictEntry(db *sql.DB, user string, dictEntry dictionaryEntry) error {
	word_id, err := findWordId(db, user, dictEntry.Word)
	if err != nil {
		log.Println("Cannot obtain word to delete", err)
		return err
	}
	_, err = db.Exec("delete from Translations where word_id = ?", word_id)
	if err != nil {
		log.Println("Error occurred while deleting dictionary entry", err)
		return err
	}
	_, err = db.Exec("delete from Idioms where word_id = ?", word_id)
	if err != nil {
		log.Println("Error occurred while deleting idioms", err)
		return err
	}
	return nil
}

func checkTranslation(db *sql.DB, user, word, translation string) (bool, error) {
	rows, err := db.Query("select 1 from Words w inner join Translations t on w.id = t.word_id where w.owner = ? and w.word = ? and t.translation = ?", user, word, translation)
	if err != nil {
		log.Println("Error occurred while checking translation", err)
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		return true, nil
	}
	return false, nil
}

func getAllWords(db *sql.DB, user string) ([]string, error) {
	rows, err := db.Query("select word from Words where owner = ?", user)
	if err != nil {
		log.Println("Error occurred while reading words", err)
		return nil, err
	}
	defer rows.Close()
	var result []string = make([]string, 0, 10)
	for rows.Next() {
		var w string
		err = rows.Scan(&w)
		if err != nil {
			log.Println("Error occurred while reading words", err)
			return nil, err
		}
		result = append(result, w)
	}
	return result, nil
}

func getDictEntry(db *sql.DB, user, word string) (*dictionaryEntry, error) {
	rows, err := db.Query("select w.word, t.translation, i.idiom from Words w inner join Translations t on w.id = t.word_id inner join Idioms i on w.id = i.word_id where w.word = ? and w.owner = ?", word, user)
	if err != nil {
		log.Println("Cannot get dictionary entry", err)
		return nil, err
	}
	defer rows.Close()
	var w, t, i string
	var entry = newDictEntry()
	for rows.Next() {
		err := rows.Scan(&w, &t, &i)
		if err != nil {
			log.Println("Cannot get dictionary entry", err)
			return nil, err
		}
		if entry.Word == "" {
			entry.Word = w
		}
		addIfAbsent(&entry.Translations, t)
		addIfAbsent(&entry.Idioms, i)
	}
	return entry, nil
}

func retrieveUsers(db *sql.DB) ([]string, error) {
	rows, err := db.Query("select distinct owner from Words")
	if err != nil {
		log.Println("Cannot retrieve users", err)
		return nil, err
	}
	defer rows.Close()
	var result = make([]string, 0, 10)
	for rows.Next() {
		var u string
		rows.Scan(&u)
		result = append(result, u)
	}
	return result, nil
}

func addIfAbsent(slice *[]string, token string) {
	for _, s := range *slice {
		if s == token {
			return
		}
	}
	*slice = append(*slice, token)
}

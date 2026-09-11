package main

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Author struct {
	ID    uint `gorm:"primaryKey"`
	Name  string
	Books []Book
}

type Book struct {
	ID       uint `gorm:"primaryKey"`
	Title    string
	AuthorID uint
}

func main() {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&Author{}, &Book{})

	db.Create(&Author{Name: "Rob Pike", Books: []Book{
		{Title: "The Go Programming Language"},
		{Title: "The Practice of Programming"},
	}})
	db.Create(&Author{Name: "Alan Donovan"})

	var count int64
	db.Model(&Author{}).Count(&count)
	fmt.Println("authors count:", count)

	var author Author
	db.Preload("Books").First(&author, "name = ?", "Rob Pike")
	fmt.Printf("author %q has %d books\n", author.Name, len(author.Books))
	for _, b := range author.Books {
		fmt.Println("  book:", b.Title)
	}

	db.Model(&Book{}).Where("title LIKE ?", "%Go%").Update("title", "The Go Programming Language (2nd)")
	var updated Book
	db.First(&updated, "author_id = ?", author.ID)
	fmt.Println("updated book:", updated.Title)

	db.Delete(&Author{}, "name = ?", "Alan Donovan")
	db.Model(&Author{}).Count(&count)
	fmt.Println("authors after delete:", count)

	var titles []string
	db.Model(&Book{}).Pluck("title", &titles)
	fmt.Println("all book titles:", titles)
}

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Book struct {
	ID        int
	Title     string
	Author    string
	Year      int
	Available int
}

type Reader struct {
	ID       int
	FullName string
}

type Loan struct {
	ID       int
	BookID   int
	ReaderID int
	Returned bool
}

type Library struct {
	Books       []Book
	Readers     []Reader
	Loans       []Loan
	BookIndex   map[int]int
	ActiveLoans map[int][]int
}

type Statistics struct {
	Titles       int
	Total        int
	Available    int
	Borrowed     int
	MostPopular  *Book
	PopularLoans int
}

func NewLibrary() *Library {
	return &Library{
		BookIndex:   make(map[int]int),
		ActiveLoans: make(map[int][]int),
	}
}

func AddBook(library *Library, book Book) error {
	if book.ID <= 0 {
		return errors.New("ID книги должен быть положительным")
	}
	if _, exists := library.BookIndex[book.ID]; exists {
		return fmt.Errorf("книга с ID %d уже существует", book.ID)
	}
	book.Title = strings.TrimSpace(book.Title)
	book.Author = strings.TrimSpace(book.Author)
	if book.Title == "" || book.Author == "" {
		return errors.New("название и автор не должны быть пустыми")
	}
	if book.Year <= 0 {
		return errors.New("год издания должен быть положительным")
	}
	if book.Available < 0 {
		return errors.New("количество экземпляров не может быть отрицательным")
	}
	library.Books = append(library.Books, book)
	library.BookIndex[book.ID] = len(library.Books) - 1
	return nil
}

func RegisterReader(library *Library, reader Reader) error {
	if reader.ID <= 0 {
		return errors.New("ID читателя должен быть положительным")
	}
	if FindReader(library.Readers, reader.ID) != nil {
		return fmt.Errorf("читатель с ID %d уже зарегистрирован", reader.ID)
	}
	reader.FullName = strings.TrimSpace(reader.FullName)
	if reader.FullName == "" {
		return errors.New("ФИО читателя не должно быть пустым")
	}
	library.Readers = append(library.Readers, reader)
	return nil
}

func FindBook(books []Book, bookIndex map[int]int, id int) *Book {
	index, exists := bookIndex[id]
	if !exists {
		return nil
	}
	return &books[index]
}

func FindReader(readers []Reader, id int) *Reader {
	for i := range readers {
		if readers[i].ID == id {
			return &readers[i]
		}
	}
	return nil
}

func FindLoan(loans []Loan, id int) *Loan {
	for i := range loans {
		if loans[i].ID == id {
			return &loans[i]
		}
	}
	return nil
}

func BorrowBook(library *Library, bookID, readerID int) error {
	if FindReader(library.Readers, readerID) == nil {
		return fmt.Errorf("читатель с ID %d не зарегистрирован", readerID)
	}
	book := FindBook(library.Books, library.BookIndex, bookID)
	if book == nil {
		return fmt.Errorf("книга с ID %d не найдена", bookID)
	}
	if book.Available <= 0 {
		return errors.New("нет доступных экземпляров этой книги")
	}
	loan := Loan{ID: len(library.Loans) + 1, BookID: bookID, ReaderID: readerID}
	library.Loans = append(library.Loans, loan)
	library.ActiveLoans[readerID] = append(library.ActiveLoans[readerID], loan.ID)
	book.Available--
	return nil
}

func ReturnBook(library *Library, bookID, readerID int) error {
	if FindReader(library.Readers, readerID) == nil {
		return fmt.Errorf("читатель с ID %d не зарегистрирован", readerID)
	}
	book := FindBook(library.Books, library.BookIndex, bookID)
	if book == nil {
		return fmt.Errorf("книга с ID %d не найдена", bookID)
	}
	active := library.ActiveLoans[readerID]
	for i, loanID := range active {
		loan := FindLoan(library.Loans, loanID)
		if loan != nil && loan.BookID == bookID && loan.ReaderID == readerID && !loan.Returned {
			loan.Returned = true
			library.ActiveLoans[readerID] = append(active[:i], active[i+1:]...)
			book.Available++
			return nil
		}
	}
	return errors.New("у этого читателя нет активной выдачи указанной книги")
}

func SearchBooks(books []Book, query string) []Book {
	query = strings.ToLower(strings.TrimSpace(query))
	var found []Book
	for _, book := range books {
		if strings.Contains(strings.ToLower(book.Title), query) || strings.Contains(strings.ToLower(book.Author), query) {
			found = append(found, book)
		}
	}
	return found
}

func GetReaderBooks(library *Library, readerID int) ([]Book, error) {
	if FindReader(library.Readers, readerID) == nil {
		return nil, fmt.Errorf("читатель с ID %d не зарегистрирован", readerID)
	}
	var books []Book
	for _, loanID := range library.ActiveLoans[readerID] {
		loan := FindLoan(library.Loans, loanID)
		if loan != nil && loan.ReaderID == readerID && !loan.Returned {
			book := FindBook(library.Books, library.BookIndex, loan.BookID)
			if book != nil {
				books = append(books, *book)
			}
		}
	}
	return books, nil
}

func CalculateStatistics(library *Library) Statistics {
	stats := Statistics{Titles: len(library.Books)}
	counts := make(map[int]int)
	for _, loan := range library.Loans {
		counts[loan.BookID]++
		if !loan.Returned {
			stats.Borrowed++
		}
	}
	for i := range library.Books {
		book := &library.Books[i]
		stats.Available += book.Available
		if counts[book.ID] > stats.PopularLoans {
			stats.MostPopular = book
			stats.PopularLoans = counts[book.ID]
		}
	}
	stats.Total = stats.Available + stats.Borrowed
	return stats
}

func WeeklyReportHeader() string {
	days := [7]string{
		"Понедельник", "Вторник", "Среда", "Четверг",
		"Пятница", "Суббота", "Воскресенье",
	}
	return "Недельный отчёт\n" + strings.Join(days[:], " | ")
}

func ReadText(scanner *bufio.Scanner, prompt string) (string, error) {
	fmt.Print(prompt)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("не удалось прочитать ввод: %w", err)
		}
		return "", io.EOF
	}
	value := strings.TrimSpace(scanner.Text())
	if value == "" {
		return "", errors.New("поле не должно быть пустым")
	}
	return value, nil
}

func ReadInt(scanner *bufio.Scanner, prompt string) (int, error) {
	text, err := ReadText(scanner, prompt)
	if err != nil {
		return 0, err
	}
	value, err := strconv.Atoi(text)
	if err != nil {
		return 0, errors.New("введите целое число")
	}
	return value, nil
}

func PrintBooks(books []Book) {
	for _, book := range books {
		fmt.Printf("ID: %d | Название: %s | Автор: %s | Год: %d | Доступно: %d\n",
			book.ID, book.Title, book.Author, book.Year, book.Available)
	}
}

func HandleChoice(library *Library, scanner *bufio.Scanner, choice int) error {
	switch choice {
	case 1:
		id, err := ReadInt(scanner, "ID книги: ")
		if err != nil {
			return err
		}
		title, err := ReadText(scanner, "Название: ")
		if err != nil {
			return err
		}
		author, err := ReadText(scanner, "Автор: ")
		if err != nil {
			return err
		}
		year, err := ReadInt(scanner, "Год издания: ")
		if err != nil {
			return err
		}
		available, err := ReadInt(scanner, "Количество доступных экземпляров: ")
		if err != nil {
			return err
		}
		if err := AddBook(library, Book{ID: id, Title: title, Author: author, Year: year, Available: available}); err != nil {
			return err
		}
		fmt.Println("Книга успешно добавлена.")
	case 2:
		id, err := ReadInt(scanner, "ID читателя: ")
		if err != nil {
			return err
		}
		name, err := ReadText(scanner, "ФИО: ")
		if err != nil {
			return err
		}
		if err := RegisterReader(library, Reader{ID: id, FullName: name}); err != nil {
			return err
		}
		fmt.Println("Читатель успешно зарегистрирован.")
	case 3, 4:
		readerID, err := ReadInt(scanner, "ID читателя: ")
		if err != nil {
			return err
		}
		bookID, err := ReadInt(scanner, "ID книги: ")
		if err != nil {
			return err
		}
		if choice == 3 {
			if err := BorrowBook(library, bookID, readerID); err != nil {
				return err
			}
			book := FindBook(library.Books, library.BookIndex, bookID)
			reader := FindReader(library.Readers, readerID)
			fmt.Printf("Книга успешно выдана!\nНазвание: %s\nЧитатель: %s\nОсталось экземпляров: %d\n",
				book.Title, reader.FullName, book.Available)
		} else {
			if err := ReturnBook(library, bookID, readerID); err != nil {
				return err
			}
			book := FindBook(library.Books, library.BookIndex, bookID)
			fmt.Printf("Книга успешно возвращена.\nНазвание: %s\nДоступно экземпляров: %d\n", book.Title, book.Available)
		}
	case 5:
		query, err := ReadText(scanner, "Введите название или автора: ")
		if err != nil {
			return err
		}
		books := SearchBooks(library.Books, query)
		if len(books) == 0 {
			fmt.Println("Книги не найдены.")
		} else {
			PrintBooks(books)
		}
	case 6:
		readerID, err := ReadInt(scanner, "ID читателя: ")
		if err != nil {
			return err
		}
		books, err := GetReaderBooks(library, readerID)
		if err != nil {
			return err
		}
		if len(books) == 0 {
			fmt.Println("У читателя нет выданных книг.")
		} else {
			PrintBooks(books)
		}
	case 7:
		stats := CalculateStatistics(library)
		fmt.Println(WeeklyReportHeader())
		fmt.Printf("Всего наименований: %d\nВсего экземпляров: %d\nДоступно: %d\nВыдано: %d\n",
			stats.Titles, stats.Total, stats.Available, stats.Borrowed)
		if stats.MostPopular == nil {
			fmt.Println("Самая популярная книга: данных о выдачах пока нет.")
		} else {
			fmt.Printf("Самая популярная книга: %s\nКоличество выдач: %d\n", stats.MostPopular.Title, stats.PopularLoans)
		}
	default:
		return errors.New("неизвестный пункт меню")
	}
	return nil
}

func main() {
	library := NewLibrary()
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("\n============ LIBRARY MANAGER ============\n\n1. Добавить книгу\n2. Зарегистрировать читателя\n3. Выдать книгу\n4. Вернуть книгу\n5. Найти книгу\n6. Книги читателя\n7. Статистика библиотеки\n0. Выход")
		choice, err := ReadInt(scanner, "Выберите действие: ")
		if err == nil {
			if choice == 0 {
				fmt.Println("Работа программы завершена.")
				return
			}
			err = HandleChoice(library, scanner, choice)
		}
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			fmt.Printf("Ошибка: %v.\n", err)
			if scanner.Err() != nil {
				return
			}
		}
	}
}

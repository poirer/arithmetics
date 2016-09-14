package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
)

type (
	// SudokuArray is a type to represent an array with 9 elements
	SudokuArray [9]int8
	// Sudoku is the whole play field
	Sudoku [9][9]int8
	// Sieve is a set of numbers allowed for a certain cell
	Sieve [81]*AllowedNumbers

	// AllowedNumbers represents a set of numbers that can be potentially written in the cell in row and col
	AllowedNumbers struct {
		lock     sync.RWMutex
		row, col int
		numbers  [9]int8
	}
)

func splitToRows(data *Sudoku) [9]SudokuArray {
	var rows [9]SudokuArray
	for i, r := range data {
		for j, n := range r {
			rows[i][j] = n
		}
	}
	return rows
}

func splitToColumns(data *Sudoku) [9]SudokuArray {
	var cols [9]SudokuArray
	for i, r := range data {
		for j, n := range r {
			cols[j][i] = n
		}
	}
	return cols
}

func splitToSquares(data *Sudoku) [9]SudokuArray {
	var squares [9]SudokuArray
	for i, r := range data {
		for j, n := range r {
			squareInd := (i/3)*3 + j/3
			cellInd := (i%3)*3 + j%3
			squares[squareInd][cellInd] = n
		}
	}
	return squares
}

func (set *SudokuArray) hasNumber(n int8) bool {
	for _, v := range set {
		if n == v {
			return true
		}
	}
	return false
}

func (set *SudokuArray) completed() bool {
	var i int8
	for i = 1; i < 10; i++ {
		if !set.hasNumber(i) {
			return false
		}
	}
	return true
}

func (an *AllowedNumbers) strikeOut(n int8) {
	an.lock.Lock()
	defer an.lock.Unlock()
	for i := 0; i < 9; i++ {
		if an.numbers[i] == n {
			an.numbers[i] = 0
			return
		}
	}
}

func fillSieve(sudoku *Sudoku) *Sieve {
	var sieve = new(Sieve)
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			if sudoku[i][j] == 0 {
				var index = i*9 + j
				sieve[index] = new(AllowedNumbers)
				sieve[index].row = i
				sieve[index].col = j
				sieve[index].numbers = fullSet
			}
		}
	}
	return sieve
}

func copySudoku(sudoku *Sudoku) *Sudoku {
	var newSudoku = new(Sudoku)
	*newSudoku = *sudoku
	return newSudoku
}

func copySieve(sieve *Sieve) *Sieve {
	var newSieve = new(Sieve)
	var index int
	for i := 0; i < 81; i++ {
		if sieve[i] != nil {
			newSieve[index] = new(AllowedNumbers)
			newSieve[index].row = sieve[i].row
			newSieve[index].numbers = sieve[i].numbers
			newSieve[index].col = sieve[i].col
			index++
		}
	}
	return newSieve
}

func parseInput(sudoku string) *Sudoku {
	var result = new(Sudoku)
	var lines = strings.Split(sudoku, "\n")
	var i, j int8
	for _, l := range lines {
		l = strings.Trim(l, " ")
		var cells = strings.Split(l, " ")
		for _, c := range cells {
			if c != "_" {
				num, err := strconv.ParseInt(c, 10, 8)
				result[i][j] = int8(num)
				if err != nil {
					log.Fatal("Invlaid input")
				}
			} else {
				result[i][j] = 0
			}
			j++
		}
		j = 0
		i++
	}
	return result
}

func PrintSudoku(sudokuData *Sudoku) {
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			if sudokuData[i][j] != 0 {
				fmt.Print(sudokuData[i][j])
			} else {
				fmt.Print("_")
			}
			fmt.Print(" ")
		}
		fmt.Println()
	}
}

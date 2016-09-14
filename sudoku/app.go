package main

import (
	"fmt"
	"sync"
)

var fullSet = SudokuArray{1, 2, 3, 4, 5, 6, 7, 8, 9}

func Sift(sieve *Sieve, rows, cols, squares [9]SudokuArray) {
	var waitGroup sync.WaitGroup
	waitGroup.Add(3)
	go siftByRows(sieve, rows, &waitGroup)
	go siftByCols(sieve, cols, &waitGroup)
	go siftBySquares(sieve, squares, &waitGroup)
	waitGroup.Wait()
}

func printAllowedNumbers(sieve *Sieve) {
	for i := 0; i < 81; i++ {
		if sieve[i] != nil && sieve[i].row == 6 && sieve[i].col == 0 {
			println("AAAAAAAAAAAAAAAAAAAA")
			fmt.Println(sieve[i].numbers)
		}
	}
}

func siftByRows(sieve *Sieve, sets [9]SudokuArray, waitGroup *sync.WaitGroup) {
	for _, entry := range sieve {
		if entry != nil {
			siftBySet(entry, &sets[entry.row])
		}
	}
	waitGroup.Done()
}

func siftByCols(sieve *Sieve, sets [9]SudokuArray, waitGroup *sync.WaitGroup) {
	for _, entry := range sieve {
		if entry != nil {
			siftBySet(entry, &sets[entry.col])
		}
	}
	waitGroup.Done()
}

func siftBySquares(sieve *Sieve, sets [9]SudokuArray, waitGroup *sync.WaitGroup) {
	for _, entry := range sieve {
		if entry != nil {
			var sqIn = (entry.row/3)*3 + entry.col/3
			siftBySet(entry, &sets[sqIn])
		}
	}
	waitGroup.Done()
}

func siftBySet(allowedNumbers *AllowedNumbers, set *SudokuArray) {
	var n int8
	for n = 1; n < 10; n++ {
		if set.hasNumber(n) {
			allowedNumbers.strikeOut(n)
		}
	}
}

func updateSudokuFromSieve(sudoku *Sudoku, sieve *Sieve) int {
	var result int
	for sieveIndex, entry := range *sieve {
		if entry != nil {
			singleNum, foundNum, noNum := defined(&entry.numbers)
			if singleNum {
				sudoku[entry.row][entry.col] = foundNum
				result++
			}
			if noNum {
				result++
				sieve[sieveIndex] = nil
			}
		}
	}
	return result
}

func defined(set *[9]int8) (bool, int8, bool) {
	var count, number int8
	for i := 0; i < 9; i++ {
		if set[i] != 0 {
			number = set[i]
			count++
		}
	}
	return count == 1, number, count == 0
}

func solved(sudoku *Sudoku) bool {
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			if sudoku[i][j] == 0 {
				return false
			}
		}
	}
	// var rows = splitToRows(sudoku)
	// var cols = splitToColumns(sudoku)
	// var sqs = splitToSquares(sudoku)
	// for i := 0; i < 9; i++ {
	// 	if !rows[i].completed() {
	// 		// println("Error in row ", i)
	// 		return false
	// 	}
	// 	if !cols[i].completed() {
	// 		println("Error in col ", i)
	// 		fmt.Println(cols[i])
	// 		return false
	// 	}
	// 	if !sqs[i].completed() {
	// 		// println("Error in square ", i)
	// 		return false
	// 	}
	// }
	return true
}

func solvedCorrectly(sudoku *Sudoku) bool {
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			if sudoku[i][j] == 0 {
				return false
			}
		}
	}
	var rows = splitToRows(sudoku)
	var cols = splitToColumns(sudoku)
	var sqs = splitToSquares(sudoku)
	for i := 0; i < 9; i++ {
		if !rows[i].completed() {
			println("Error in row ", i)
			return false
		}
		if !cols[i].completed() {
			println("Error in col ", i)
			fmt.Println(cols[i])
			return false
		}
		if !sqs[i].completed() {
			println("Error in square ", i)
			return false
		}
	}
	return true
}

// This function sifts numbers until the count of updated  cells is zero.
// It is possible that the function will work concurrently so it takes stop channel
// When a goroutine solves sudoku, it writes something in the channel and all concurrent calls interrupts
// Flag is written back to chan to stope other functions if any
func applySieve(sudoku *Sudoku, sieve *Sieve, stopChan chan bool) {
	var waitGroup sync.WaitGroup
	for {
		select {
		case stopFlag := <-stopChan:
			println("Stop flag received")
			stopChan <- stopFlag
			return
		default:
		}
		waitGroup.Add(3)
		var rows, cols, sqs [9]SudokuArray
		go func() {
			rows = splitToRows(sudoku)
			waitGroup.Done()
		}()
		go func() {
			cols = splitToColumns(sudoku)
			waitGroup.Done()
		}()
		go func() {
			sqs = splitToSquares(sudoku)
			waitGroup.Done()
		}()
		waitGroup.Wait()
		Sift(sieve, rows, cols, sqs)
		if updateSudokuFromSieve(sudoku, sieve) == 0 {
			break
		}
	}
}

func eliminateNUmbersAfterSuggestion(sieve *Sieve, row, col int) {
	for j := 0; j < 81; j++ {
		if sieve[j] != nil && sieve[j].row == row && sieve[j].col == col {
			sieve[j] = nil
			break
		}
	}
}

// tryToSolveWithSuggestions makes attempts to solve sudoku using suggestions. If a cell may contain 2 or more numbers,
// then it tries to set one of the allowed numbers and then apply solution algorithm. It may lead to a solution.
func tryToSolveWithSuggestions(sudoku *Sudoku, sieve *Sieve, indexToSuggest int, channelToAnswer chan *Sudoku, stopChan chan bool) {
	var allowedNumbers = sieve[indexToSuggest]
	for i := 0; i < 9; i++ {
		if allowedNumbers.numbers[i] != 0 {
			var sudokuForSuggestion = copySudoku(sudoku)
			var sieveForSuggestion = copySieve(sieve)
			sudokuForSuggestion[allowedNumbers.row][allowedNumbers.col] = allowedNumbers.numbers[i]
			eliminateNUmbersAfterSuggestion(sieveForSuggestion, allowedNumbers.row, allowedNumbers.col)
			applySieve(sudokuForSuggestion, sieveForSuggestion, stopChan)
			if solvedCorrectly(sudokuForSuggestion) {
				channelToAnswer <- sudokuForSuggestion
				stopChan <- true // Remove this line if we want to get all solutions, not just the first one
				break
			}
		}
	}
	channelToAnswer <- nil
}

func main() {
	var sudokuInput = `1 _ 3 _ _ 6 _ 8 _
  _ 5 _ _ 8 _ 1 2 _
  7 _ 9 1 _ 3 _ 5 6
  _ 3 _ _ 6 7 _ 9 _
  5 _ 7 8 _ _ _ 3 _
  8 _ 1 _ 3 _ 5 _ 7
  _ 4 _ _ 7 8 _ 1 _
  6 _ 8 _ _ 2 _ 4 _
  _ 1 2 _ 4 5 _ 7 8`

	var sudokuData = parseInput(sudokuInput)
	var sieve = fillSieve(sudokuData)
	var stopChan = make(chan bool, 81)
	applySieve(sudokuData, sieve, stopChan)
	if !solved(sudokuData) {
		var expectedResultCount int
		for i := 0; i < 81; i++ {
			if sieve[i] != nil {
				expectedResultCount++
			}
		}
		// If strict algorithm didn't give a solution, then we try to populate cells with non-sifted numbers and see if after that solution can be found
		// All attempts are being done in separate goroutine; then we wait for all of them to see whatsuggestions were right and what were wromg.
		// It may make sense to wait for the first solution; then we need just return from the main function
		var resChan = make(chan *Sudoku, expectedResultCount)
		for i := 0; i < 81; i++ {
			if sieve[i] != nil {
				go tryToSolveWithSuggestions(sudokuData, sieve, i, resChan, stopChan)
			}
		}
		for i := 0; i < expectedResultCount; i++ {
			select {
			case s := <-stopChan:
				stopChan <- s // to skip reading of results
			case s := <-resChan:
				if s == nil {
					println("Suggestion was wrong")
				} else {
					println("Solved by making suggestion")
					PrintSudoku(s)
				}
			}
		}
	} else {
		println("Solved using straight way")
		PrintSudoku(sudokuData)
	}
}

package main

import "testing"

const testSudokuInput = `1 _ 3 _ _ 6 _ 8 _
  _ 5 _ _ 8 _ 1 2 _
  7 _ 9 1 _ 3 _ 5 6
  _ 3 _ _ 6 7 _ 9 _
  5 _ 7 8 _ _ _ 3 _
  8 _ 1 _ 3 _ 5 _ 7
  _ 4 _ _ 7 8 _ 1 _
  6 _ 8 _ _ 2 _ 4 _
  _ 1 2 _ 4 5 _ 7 8`

var (
	sudokuData = Sudoku{
		{1, 0, 3, 0, 0, 6, 0, 8, 0},
		{0, 5, 0, 0, 8, 0, 1, 2, 0},
		{7, 0, 9, 1, 0, 3, 0, 5, 6},
		{0, 3, 0, 0, 6, 7, 0, 9, 0},
		{5, 0, 7, 8, 0, 0, 0, 3, 0},
		{8, 0, 1, 0, 3, 0, 5, 0, 7},
		{0, 4, 0, 0, 7, 8, 0, 1, 0},
		{6, 0, 8, 0, 0, 2, 0, 4, 0},
		{0, 1, 2, 0, 4, 5, 0, 7, 8},
	}

	sudokuRows = [9]SudokuArray{
		{1, 0, 3, 0, 0, 6, 0, 8, 0},
		{0, 5, 0, 0, 8, 0, 1, 2, 0},
		{7, 0, 9, 1, 0, 3, 0, 5, 6},
		{0, 3, 0, 0, 6, 7, 0, 9, 0},
		{5, 0, 7, 8, 0, 0, 0, 3, 0},
		{8, 0, 1, 0, 3, 0, 5, 0, 7},
		{0, 4, 0, 0, 7, 8, 0, 1, 0},
		{6, 0, 8, 0, 0, 2, 0, 4, 0},
		{0, 1, 2, 0, 4, 5, 0, 7, 8},
	}

	sudokuCols = [9]SudokuArray{
		{1, 0, 7, 0, 5, 8, 0, 6, 0},
		{0, 5, 0, 3, 0, 0, 4, 0, 1},
		{3, 0, 9, 0, 7, 1, 0, 8, 2},
		{0, 0, 1, 0, 8, 0, 0, 0, 0},
		{0, 8, 0, 6, 0, 3, 7, 0, 4},
		{6, 0, 3, 7, 0, 0, 8, 2, 5},
		{0, 1, 0, 0, 0, 5, 0, 0, 0},
		{8, 2, 5, 9, 3, 0, 1, 4, 7},
		{0, 0, 6, 0, 0, 7, 0, 0, 8},
	}

	sudokuSquares = [9]SudokuArray{
		{1, 0, 3, 0, 5, 0, 7, 0, 9},
		{0, 0, 6, 0, 8, 0, 1, 0, 3},
		{0, 8, 0, 1, 2, 0, 0, 5, 6},
		{0, 3, 0, 5, 0, 7, 8, 0, 1},
		{0, 6, 7, 8, 0, 0, 0, 3, 0},
		{0, 9, 0, 0, 3, 0, 5, 0, 7},
		{0, 4, 0, 6, 0, 8, 0, 1, 2},
		{0, 7, 8, 0, 0, 2, 0, 4, 5},
		{0, 1, 0, 0, 4, 0, 0, 7, 8},
	}
)

func TestParseInput(t *testing.T) {
	var parseResult = parseInput(testSudokuInput)
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			if parseResult[i][j] != sudokuData[i][j] {
				t.Errorf("Invalid parse result at %d:%d. Expected %d but was %d", i, j, sudokuData[i][j], parseResult[i][j])
			}
		}
	}
}

func TestSplitToRows(t *testing.T) {
	checkSplit("TestSplitToRows", t, sudokuRows, splitToRows)
}

func TestSplitToCols(t *testing.T) {
	checkSplit("TestSplitToCols", t, sudokuCols, splitToColumns)
}

func TestSplitToSquares(t *testing.T) {
	checkSplit("TestSplitToSquares", t, sudokuSquares, splitToSquares)
}

func checkSplit(name string, t *testing.T, expectedData [9]SudokuArray, splitFunc func(*Sudoku) [9]SudokuArray) {
	var splitResult = splitFunc(&sudokuData)
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			if splitResult[i][j] != expectedData[i][j] {
				t.Errorf("Split result does not match to expected at %d:%d. Expected %d but was %d", i, j, expectedData[i][j], splitResult[i][j])
			}
		}
	}
}

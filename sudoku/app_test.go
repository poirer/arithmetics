package main

import "testing"

func TestSolveSudoku(t *testing.T) {
	var sudokuInput = `1 _ 3 _ _ 6 _ 8 _
  _ 5 _ _ 8 _ 1 2 _
  7 _ 9 1 _ 3 _ 5 6
  _ 3 _ _ 6 7 _ 9 _
  5 _ 7 8 _ _ _ 3 _
  8 _ 1 _ 3 _ 5 _ 7
  _ 4 _ _ 7 8 _ 1 _
  6 _ 8 _ _ 2 _ 4 _
  _ 1 2 _ 4 5 _ 7 8`

	var expectedResult = []Sudoku{
		{{1, 2, 3, 4, 5, 6, 7, 8, 9},
			{4, 5, 6, 7, 8, 9, 1, 2, 3},
			{7, 8, 9, 1, 2, 3, 4, 5, 6},
			{2, 3, 4, 5, 6, 7, 8, 9, 1},
			{5, 6, 7, 8, 9, 1, 2, 3, 4},
			{8, 9, 1, 2, 3, 4, 5, 6, 7},
			{3, 4, 5, 9, 7, 8, 6, 1, 2},
			{6, 7, 8, 3, 1, 2, 9, 4, 5},
			{9, 1, 2, 6, 4, 5, 3, 7, 8}},
		{{1, 2, 3, 4, 5, 6, 7, 8, 9},
			{4, 5, 6, 7, 8, 9, 1, 2, 3},
			{7, 8, 9, 1, 2, 3, 4, 5, 6},
			{2, 3, 4, 5, 6, 7, 8, 9, 1},
			{5, 6, 7, 8, 9, 1, 2, 3, 4},
			{8, 9, 1, 2, 3, 4, 5, 6, 7},
			{3, 4, 5, 6, 7, 8, 9, 1, 2},
			{6, 7, 8, 9, 1, 2, 3, 4, 5},
			{9, 1, 2, 3, 4, 5, 6, 7, 8}},
		{{1, 2, 3, 4, 5, 6, 7, 8, 9},
			{4, 5, 6, 7, 8, 9, 1, 2, 3},
			{7, 8, 9, 1, 2, 3, 4, 5, 6},
			{2, 3, 4, 5, 6, 7, 8, 9, 1},
			{5, 6, 7, 8, 9, 1, 2, 3, 4},
			{8, 9, 1, 2, 3, 4, 5, 6, 7},
			{9, 4, 5, 6, 7, 8, 3, 1, 2},
			{6, 7, 8, 3, 1, 2, 9, 4, 5},
			{3, 1, 2, 9, 4, 5, 6, 7, 8}},
		{{1, 2, 3, 4, 5, 6, 7, 8, 9},
			{4, 5, 6, 7, 8, 9, 1, 2, 3},
			{7, 8, 9, 1, 2, 3, 4, 5, 6},
			{2, 3, 4, 5, 6, 7, 8, 9, 1},
			{5, 6, 7, 8, 9, 1, 2, 3, 4},
			{8, 9, 1, 2, 3, 4, 5, 6, 7},
			{9, 4, 5, 3, 7, 8, 6, 1, 2},
			{6, 7, 8, 9, 1, 2, 3, 4, 5},
			{3, 1, 2, 6, 4, 5, 9, 7, 8}},
	}

	var sudokuData = parseInput(sudokuInput)
	res, _ := solveSudoku(sudokuData, true)
	if len(res) != 4 {
		t.Error("Wrong number of found solutions")
	}
	var foundFlag int
	for i := 0; i < len(res); i++ {
		var p = findMatchedSolution(res[i], &expectedResult)
		foundFlag |= 1 << uint(p)
	}
	if foundFlag != 0xf {
		t.Error("Found set of solution is wrong")
	}
	sudokuData = parseInput(sudokuInput)
	res, _ = solveSudoku(sudokuData, false)
	if len(res) != 1 {
		t.Error("Wrong number of found solutions")
		t.Fail()
	}
	if findMatchedSolution(res[0], &expectedResult) == -1 {
		t.Error("Unexpected solution")
	}
}

func findMatchedSolution(s *Sudoku, allResults *[]Sudoku) int {
	for i := 0; i < len(*allResults); i++ {
		if sudokuAreEqual(s, &(*allResults)[i]) {
			return i
		}
	}
	return -1
}

func sudokuAreEqual(s1, s2 *Sudoku) bool {
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			if s1[i][j] != s2[i][j] {
				return false
			}
		}
	}
	return true
}

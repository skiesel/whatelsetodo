# WhatElseToDo

Simple command line utility for reading a project for tagged comments

Sample output from running:



```
> whatslefttodo -dir="test"
test/test1.c
	FIX:
		1: line 1
		7: line 7
		11: line 11
		14: line 14
		16: line 16 FIX: line 16
	TODO:
		24: 24
test/test2.c
	FIX:
		1: line 1
		7: line 7
		11: line 11
		14: line 14
		16: line 16 FIX: line 16
test/inner/test.c
	FIX:
		1: line 1
		7: line 7
		11: line 11
		14: line 14
		16: line 16 FIX: line 16
test/inner/test.h
	FIX:
		1: line 1
		7: line 7
		11: line 11
		14: line 14
		16: line 16 FIX: line 16
```

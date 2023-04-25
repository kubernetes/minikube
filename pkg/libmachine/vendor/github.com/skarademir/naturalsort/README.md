# naturalsort
A simple natural string sorter for Go.

##Usage
Implements the `sort.Interface`

called by `sort.Sort(NaturalSort([]string))`
###Example

```go
SampleStringArray := []string{
                       "z24", "z2", "z15", "z1",
                       "z3", "z20", "z5", "z11",
                       "z 21", "z22"}
sort.Sort(NaturalSort(SampleStringArray))
```

##Needless Description
Inspired by [Jeff Atwood's seminal blog post](http://blog.codinghorror.com/sorting-for-humans-natural-sort-order/) and 
structured similarly to [Ian Griffiths' C# implementation](http://www.interact-sw.co.uk/iangblog/2007/12/13/natural-sorting).
This uses a regex to split the numeric and non-numeric portions of the string into a chunky array. Next, the left and right sides'
chunks are compared either by string comparrison (if either chunk is a non-numeric), or by ~~integer (if both chunks are numeric)~~ a character-by-character iterative function that compares numerical strings

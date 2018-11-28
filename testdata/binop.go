package main

func main() {
	println("string equality")
	println(a == "a")
	println(a == "b")
	println(a != "a")
	println(a != "b")
	println("string inequality")
	println(a > "b")
	println("b" > a)
	println("b" > "b")
	println(a <= "b")
	println("b" <= a)
	println("b" <= "b")
	println(a > "b")
	println("b" > a)
	println("b" > "b")
	println(a <= "b")
	println("b" <= a)
	println("b" <= "b")
	println(a < "aa")
	println("aa" < a)
	println("ab" < "aa")
	println("aa" < "ab")

	println("array equality")
	println(a1 == [2]int{1, 2})
	println(a1 != [2]int{1, 2})
	println(a1 == [2]int{1, 3})
	println(a1 == [2]int{2, 2})
	println(a1 == [2]int{2, 1})
	println(a1 != [2]int{2, 1})

	println("struct equality")
	println(s1 == Struct1{3, true})
	println(s1 == Struct1{4, true})
	println(s1 == Struct1{3, false})
	println(s1 == Struct1{4, false})
	println(s1 != Struct1{3, true})
	println(s1 != Struct1{4, true})
	println(s1 != Struct1{3, false})
	println(s1 != Struct1{4, false})

	println("blank fields in structs")
	println(s2 == Struct2{"foo", 0.0, 5})
	println(s2 == Struct2{"foo", 0.0, 7})
	println(s2 == Struct2{"foo", 1.0, 5})
	println(s2 == Struct2{"foo", 1.0, 7})
}

var x = true
var y = false

var a = "a"
var s1 = Struct1{3, true}
var s2 = Struct2{"foo", 0.0, 5}

var a1 = [2]int{1, 2}

type Int int

type Struct1 struct {
	i Int
	b bool
}

type Struct2 struct {
	s string
	_ float64
	i int
}

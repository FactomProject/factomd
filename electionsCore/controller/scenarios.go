package controller

// Vol  To
//  1 { 1 2 }<-v
//		Route vol 1 to 1, and 2

//  From    Vote    To
// { 1 2 }   1    { 0 2 } <-o
//		Route vote 1 from (0, 2) to (1, 2)

//  From   Level    To
// { 1 2 }   1    { 0 2 } <-o
//		Route level 1 from (0, 2) to (1, 2)

var Scenarios = make(map[string]string)

func init() {
	Scenarios["flipfloph"] = HorizontalFlipflop
	Scenarios["agree"] = AllAgree
}

// " flipfloph" runscene

var HorizontalFlipflop string = `
# Setup some variables
{ [ 0 1 2 ] } /all def
{ [ 0 1 ] }   /left def
{ [ 1 2 ] }   /right def
{ [ 1 ] }     /mid def
{ [ 2 ] }     /fright def

# Pass around volunteer 1
1 all <-v

# Pass around volunteer 2
2 all <-v

# Voting time
left 1 left <-o
right 2 right <-o

# /* levels */
left 1 left <-l
mid 2 right <-l
fright 1 mid <-l

# Turn on routing
<r>

{ 2 } 1 { 0 } <-l
{ 1 } 2 { 0 } <-l
{ 1 } 4 { 0 } <-l
{ 2 } 2 { 0 } <-l

.ca
`

var AllAgree string = `
32 1 setcon
# Setup some variables

{ [ 0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31 ] } /all def

# Pass around volunteer 1
0 all <-v
all 0 all <-o


# /* levels */
all 1 all <-l
all 2 all <-l
all 3 all <-l
all 4 all <-l
all 5 all <-l
all 6 all <-l

.ca
`

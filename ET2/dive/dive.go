package dive

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/FactomProject/electiontesting/controller"
	"github.com/FactomProject/electiontesting/election"
	"crypto/sha256"

	"time"

	"flag"

	"github.com/FactomProject/electiontesting/messages"
	"github.com/FactomProject/electiontesting/primitives"
	"github.com/dustin/go-humanize"
	. "github.com/FactomProject/electiontesting/ET2/mirrors"
	. "github.com/FactomProject/electiontesting/ET2/directedmessage"
)

var MirrorMap Mirrors

var solutions = 0
var breadth = 0
var loops = 0
var mirrors = 0
var depths []int
var solutionsAt []int
var mirrorsAt []int
var deadMessagesAt []int
var failuresAt []int
var hitlimit int
var maxdepth int
var failure int
var errCollision int
var winners []int

var globalRunNumber = 0

var leadersMap = make(map[primitives.Identity]int)
var audsMap = make(map[primitives.Identity]int)

var extraPrints = true
var extraPrints1 = true
var extraPrints2 = true
var extraPrints3 = true
var insanePrints = false

var last time.Time
var primelist = []int{101399, 101411, 101419, 101429, 101449, 101467, 101477, 101483, 101489, 101501,
	101503, 101513, 101527, 101531, 101533, 101537, 101561, 101573, 101581, 101599,
	101603, 101611, 101627, 101641, 101653, 101663, 101681, 101693, 101701, 101719,
	101723, 101737, 101741, 101747, 101749, 101771, 101789, 101797, 101807, 101833,
	101837, 101839, 101863, 101869, 101873, 101879, 101891, 101917, 101921, 101929,
	101939, 101957, 101963, 101977, 101987, 101999, 102001, 102013, 102019, 102023,
	102031, 102043, 102059, 102061, 102071, 102077, 102079, 102101, 102103, 102107,
	102121, 102139, 102149, 102161, 102181, 102191, 102197, 102199, 102203, 102217,
	102229, 102233, 102241, 102251, 102253, 102259, 102293, 102299, 102301, 102317,
	102329, 102337, 102359, 102367, 102397, 102407, 102409, 102433, 102437, 102451,
	102461, 102481, 102497, 102499, 102503, 102523, 102533, 102539, 102547, 102551,
	102559, 102563, 102587, 102593, 102607, 102611, 102643, 102647, 102653, 102667,
	102673, 102677, 102679, 102701, 102761, 102763, 102769, 102793, 102797, 102811,
	102829, 102841, 102859, 102871, 102877, 102881, 102911, 102913, 102929, 102931,
	102953, 102967, 102983, 103001, 103007, 103043, 103049, 103067, 103069, 103079,
	103087, 103091, 103093, 103099, 103123, 103141, 103171, 103177, 103183, 103217,
	103231, 103237, 103289, 103291, 103307, 103319, 103333, 103349, 103357, 103387,
	103391, 103393, 103399, 103409, 103421, 103423, 103451, 103457, 103471, 103483,
	103511, 103529, 103549, 103553, 103561, 103567, 103573, 103577, 103583, 103591,
	103613, 103619, 103643, 103651, 103657, 103669, 103681, 103687, 103699, 103703,
	103723, 103769, 103787, 103801, 103811, 103813, 103837, 103841, 103843, 103867,
	103889, 103903, 103913, 103919, 103951, 103963, 103967, 103969, 103979, 103981,
	103991, 103993, 103997, 104003, 104009, 104021, 104033, 104047, 104053, 104059,
	104087, 104089, 104107, 104113, 104119, 104123, 104147, 104149, 104161, 104173,
	104179, 104183, 104207, 104231, 104233, 104239, 104243, 104281, 104287, 104297,
	104309, 104311, 104323, 104327, 104347, 104369, 104381, 104383, 104393, 104399,
	104417, 104459, 104471, 104473, 104479, 104491, 104513, 104527, 104537, 104543,
	104549, 104551, 104561, 104579, 104593, 104597, 104623, 104639, 104651, 104659,
	104677, 104681, 104683, 104693, 104701, 104707, 104711, 104717, 104723, 104729}

var recursions, randomFactor, primeIdx, global int

//================ main =================
func Main() {
	audits := flag.Int("a", 2, "Number of audit servers")
	feds := flag.Int("f", 3, "Number of federated servers")
	recursionsPtr := flag.Int("r", 1000, "Number of recursions allowed")
	randomFactorPtr := flag.Int("p", 1, "Pick a starting prime")
	globalPtr := flag.Int("g", 1000, "How many global nodes between prints")
	flag.Parse()

	primeIdx = *randomFactorPtr
	global = *globalPtr
	recursions = *recursionsPtr

	fmt.Println("Settings:")
	fmt.Println("  a Audits:       ", audits)
	fmt.Println("  f Feds:         ", feds)
	fmt.Println("  r Recursions:   ", recursions)
	fmt.Println("  p PrimeFactor:  ", randomFactor)
	fmt.Println("  g Global int:   ", global)
	recurse(*audits, *feds, recursions)
}

func SetGlobals(r int, rf int, p int, g int) {
	recursions, randomFactor, primeIdx, global = r, rf, p, g
}

// newElections will return an array of elections (1 per leader) and an array
// of volunteers messages to kick things off.
//		Params:
//			feds int   Number of Federated Nodes
//			auds int   Number of Volunteers
//			noDisplay  Passing a true here will reduce memory consumption, as it is a debugging tool
//
//		Returns:
//			controller *Controller  This can used for debugging (Printing votes)
//			elections []*election   Nodes you can execute on (returns msg, statchange)
//			volmsgs   []*VoluntMsg	Volunteer msgs you can start things with
func newElections(feds, auds int, noDisplay bool) (*controller.Controller, []*election.Election, []*DirectedMessage) {
	con := controller.NewController(feds, auds)

	if noDisplay {
		for _, e := range con.Elections {
			e.Display = nil
		}
		con.GlobalDisplay = nil
	}
	var msgs []*DirectedMessage
	fmt.Println("Starting")
	for _, v := range con.Volunteers {
		for i, _ := range con.Elections {
			my := new(DirectedMessage)
			my.LeaderIdx = i
			my.Msg = v
			msgs = append(msgs, my)
			fmt.Println(my.Msg.String(), my.LeaderIdx)
		}
	}
	{
		global := con.Elections[0].Display.Global
		for i, ldr := range con.Elections {
			con.Elections[i] = CloneElection(ldr)
			con.Elections[i].Display.Global = global
		}
	}
	for _, l := range con.Elections {
		leadersMap[l.Self] = con.Elections[0].FedIDtoIndex(l.Self)
	}

	for _, a := range con.AuthSet.GetAuds() {
		audsMap[a] = con.Elections[0].GetVolunteerPriority(a)
	}

	return con, con.Elections, msgs
}

var cnt = 0

// dive
// Pass a list of messages to process, to a set of leaders, at a current depth, with a particular limit.
// Provided a msgPath, and updated for recording purposes.
// Returns
// limitHit -- path hit the limit, and recursed.  Can happen on loops
// leaf -- All messages were processed, and no message resulted in a change of state.
// seeSuccess -- Some path below this dive produced a solution
// Note that we actually dive 100 levels beyond our limit, and declare seeSuccess past our limit as proof we are
// in a loop.
// Hitting the limit and seeSuccess is proof of a loop that none the less can resolve.
func Dive(mList []*DirectedMessage, leaders []*election.Election, depth int, limit int, msgPath []*DirectedMessage) (limitHit bool, leaf bool, seeSuccess bool) {
	depths = incCounter(depths, depth)
	depth++

	now := time.Now()

	globalRunNumber++
	if globalRunNumber%global == 0 {
		//printState(depth, mList, leaders, msgPath)
		extraPrints = true
		extraPrints1 = true
		extraPrints2 = true
		extraPrints3 = true
		last = now
	}

	if depth > limit {
		if extraPrints {
			fmt.Println("////////>>>>>>>>> Hit Limit <<<<<<<<<<<<<<<<<")
			printState(depth, mList, leaders, msgPath)
			extraPrints = false
		}
		breadth++
		hitlimit++
		return true, false, false
	}

	if depth > maxdepth {
		maxdepth = depth
	}

	if depth < 2 {
		fmt.Println("////////========== Depth %d ====================")
		printState(depth, mList, leaders, msgPath)
	}

	//done := 0
	//for _, ldr := range leaders {
	//	if ldr.Committed {
	//		done++
	//	}
	//}
	if complete, err := nodesCompleted(leaders); complete { // done == len(leaders)/2+1 {
		solutionsAt = incCounter(solutionsAt, depth)
		if extraPrints {
			//			fmt.Println(">>>>>>>>>>>>>>>>>>>>> Solution Found @ ", depth)
		}
		for _, ldr := range leaders {
			if ldr.Committed {
				win := int(ldr.CurrentVote.VolunteerPriority)
				for len(winners) <= win {
					winners = append(winners, 0)
				}
				winners[win]++
				break
			}
		}
		breadth++
		solutions++
		if extraPrints2 {
			fmt.Println("////////!!!!!!!  Success!")
			printState(depth, mList, leaders, msgPath)
			extraPrints2 = false
		}
		return false, true, true

	} else if err != nil {
		// Bad! This means the algorithm is broken
		errCollision++
		if extraPrints3 {
			extraPrints3 = false
			fmt.Println("// Fail with collision @ depth:", depth)
			fmt.Println(err.Error())
			printState(depth, mList, leaders, msgPath)
		}
	}

	// Look for MirrorMap, but only after we have been going a bit.
	if depth > 4 {
		var hashes [][32]byte
		var strings []string
		for _, ldr := range leaders {
			bits := ldr.NormalizedString()
			if bits != nil {
				strings = append(strings, string(bits))
				h := Sha(bits)
				hashes = append(hashes, h)
			} else {
				panic("shouldn't happen")
			}
		}
		for i := 0; i < len(hashes)-1; i++ {
			for j := 0; j < len(hashes)-1-i; j++ {
				if bytes.Compare(hashes[j][:], hashes[j+1][:]) > 0 {
					hashes[j], hashes[j+1] = hashes[j+1], hashes[j+1]
					strings[j], strings[j+1] = strings[j+1], strings[j]
				}
			}
		}
		var all []byte
		var alls string
		for i, h := range hashes {
			all = append(all, h[:]...)
			alls += strings[i]
		}
		mh := Sha(all)
		if MirrorMap.IsMirror(mh) {
			mirrors++
			breadth++
			mirrorsAt = incCounter(mirrorsAt, depth)
			return false, false, true
		}
	}

	leaf = true

	shuffle := make([]*DirectedMessage, len(mList))
	copy(shuffle, mList)

	i := primeIdx % len(shuffle)
	d := primelist[primeIdx%len(primelist)]
	primeIdx += d
	j := d % len(shuffle)
	shuffle[i], shuffle[j] = shuffle[j], shuffle[i]
	didx := primelist[primeIdx%len(primelist)]
	didi := primeIdx
	for _, v := range shuffle {
		d := didi % len(shuffle)
		didi += didx

		var msgs2 []*DirectedMessage
		msgs2 = append(msgs2, shuffle[0:d]...)
		msgs2 = append(msgs2, shuffle[d+1:]...)
		ml2 := len(msgs2)

		cl := CloneElection(leaders[v.LeaderIdx])

		//if !spewSame(cl, leaders[v.LeaderIdx]) {
		//	fmt.Println("Clone Failed")
		//	debugClone(cl, leaders[v.LeaderIdx])
		//	os.Exit(0)
		//}

		if leaders[v.LeaderIdx].Committed {
			continue
		}

		msg, changed := leaders[v.LeaderIdx].Execute(v.Msg, depth)
		hprime := leaders[v.LeaderIdx].StateString()
		for i := 0; i < 10; i++ {

			c2 := CloneElection(cl)
			c2.Execute(v.Msg, depth)
			hclone := c2.StateString()

			if bytes.Compare(hprime, hclone) != 0 {
				fmt.Println("\nsending: ", formatForInterpreter(v))
				fmt.Println("-----------------------")
				fmt.Println(cl.Display.String())
				fmt.Println(len(hprime), string(hprime))
				fmt.Println(c2.Display.String())
				fmt.Println(len(hclone), string(hclone))
				printState(depth, msgs2, leaders, msgPath)
				time.Sleep(2 * time.Second)
				panic("Ekk")
			} else {
				//fmt.Print("+")
			}
		}

		msgPath2 := append(msgPath, v)

		if changed {
			leaf = false
			if msg != nil {
				for i, _ := range leaders {
					if i != v.LeaderIdx {
						my := new(DirectedMessage)
						my.LeaderIdx = i
						my.Msg = msg
						msgs2 = append(msgs2, my)
					}
				}
			}
			gl := leaders[v.LeaderIdx].Display.Global
			for _, ldr := range leaders {
				ldr.Display.Global = gl
			}
			// Recursive Dive
			lim, _, ss := Dive(msgs2, leaders, depth, limit, msgPath2)
			_ = lim || ss
			seeSuccess = seeSuccess || ss
			limitHit = limitHit || lim
			for _, ldr := range leaders {
				ldr.Display.Global = cl.Display.Global
			}
			msgs2 = msgs2[:ml2]
		} else {
			deadMessagesAt = incCounter(deadMessagesAt, depth)
		}
		leaders[v.LeaderIdx] = cl
	}
	if limitHit {
		leaf = false
	}
	if limitHit {
		if depth == 9 {

			if seeSuccess {
				loops++
			} else {
				failure++
				if extraPrints1 {
					extraPrints1 = false
					fmt.Println("////////>>>>>>>/// Loops Fail //////////////////////")
					fmt.Printf("%d %d setcon\n", len(leadersMap), len(audsMap))
					printState(depth, shuffle, leaders, msgPath)
				}
			}
			limitHit = false
		}
	} else {
		if leaf {
			incCounter(failuresAt, depth)
			failure++
			leaf = false

			if extraPrints1 {
				extraPrints1 = false
				fmt.Println("/////////////// Fail //////////////////////")
				fmt.Printf("%d %d setcon\n", len(leadersMap), len(audsMap))
				printState(depth, shuffle, leaders, msgPath)
			}

		}
	}

	return limitHit, leaf, seeSuccess
}

func printState(depth int, msgs []*DirectedMessage, leaders []*election.Election, msgPath []*DirectedMessage) {
	fmt.Printf("%s%s%4d%s%4d %s %12s %s%12s %s%5d %s%12s %12s  %s %12s %s %12s %s %12s %s %12s %s %12s", "=============== ",
		" Depth=", depth, "/", maxdepth,
		"| Multiple Collision", humanize.Comma(int64(errCollision)),
		"| Failures=", humanize.Comma(int64(failure)),
		"| MsgQ=", len(msgs),
		"| Mirrors=", humanize.Comma(int64(mirrors)), humanize.Comma(int64(MirrorMap.Len())),
		"| Hit the Limits=", humanize.Comma(int64(hitlimit)),
		"| Breadth=", humanize.Comma(int64(breadth)),
		"| solutions so far =", humanize.Comma(int64(solutions)),
		"| global count= ", humanize.Comma(int64(globalRunNumber)),
		"| loops detected=", humanize.Comma(int64(loops)))

	prt := func(counter []int, msg string) {
		fmt.Printf("\n=%20s", msg)
		if len(counter) == 0 {
			fmt.Println("\n=     None Found\n=")
		}
		for i, v := range counter {
			if i%16 == 0 {
				fmt.Println("")
				fmt.Print("=")
			}
			str := fmt.Sprintf("%s[%3d]", humanize.Comma(int64(v)), i)
			fmt.Printf("%12s ", str)
		}
	}
	prt(winners, "Winning Volunteers")
	prt(deadMessagesAt, "Dead Messages")
	prt(mirrorsAt, "Mirrors")
	prt(solutionsAt, "Solutions")
	prt(failuresAt, "Failures")
	prt(depths, "Depths")
	fmt.Println()

	fmt.Println("====Start List===")
	// Lots of printing... Not necessary....
	fmt.Println(leaders[0].Display.Global.String())
	fmt.Println("====End Global===")

	for _, ldr := range leaders {
		fmt.Println(ldr.Display.String())
	}
	fmt.Println("====End List===")

	if insanePrints {
		// Example of a run that has a werid msg state
		fmt.Println("Leader 0")
		fmt.Println(leaders[0].PrintMessages())
		fmt.Println("Leader 1")
		fmt.Println(leaders[1].PrintMessages())
		fmt.Println("Leader 2")
		fmt.Println(leaders[2].PrintMessages())
	}
	fmt.Printf("%d %d setcon\n", len(leadersMap), len(audsMap))
	for i, v := range msgPath {
		fmt.Println(formatForInterpreter(v), "#", i, v.LeaderIdx, "<==", leaders[0].Display.FormatMessage(v.Msg))
	}
	fmt.Println("<b> # Pending:")
	for i, v := range msgs {
		fmt.Println(formatForInterpreter(v), "#", i, v.LeaderIdx, "<==", leaders[0].Display.FormatMessage(v.Msg))
	}

}

func nodesCompleted(nodes []*election.Election) (bool, error) {
	done := 0
	prev := -1
	for _, n := range nodes {
		if n.Committed {
			done++
			if prev != -1 && n.CurrentVote.VolunteerPriority != prev {
				return false, fmt.Errorf("2 nodes committed on different results. %d and %d", prev, n.CurrentVote.VolunteerPriority)
			}
			prev = n.CurrentVote.VolunteerPriority
		}
	}

	return done >= len(nodes)/2+1, nil
}

func formatForInterpreter(my *DirectedMessage) string {
	msg := my.Msg
	switch msg.(type) {
	case *messages.LeaderLevelMessage:
		l := msg.(*messages.LeaderLevelMessage)
		from := leadersMap[l.Signer]

		return fmt.Sprintf("{ %d } %d { %d } <-l", from, l.Level, my.LeaderIdx)
	case *messages.VolunteerMessage:
		a := msg.(*messages.VolunteerMessage)
		from := audsMap[a.Signer]

		return fmt.Sprintf("%d { %d } <-v", from, my.LeaderIdx)
	case *messages.VoteMessage:
		l := msg.(*messages.VoteMessage)
		from := leadersMap[l.Signer]
		vol := audsMap[l.Volunteer.Signer]

		return fmt.Sprintf("{ %d } %d { %d } <-o", from, vol, my.LeaderIdx)
	}
	return "NA"
}

func incCounter(counter []int, depth int) []int {
	for len(counter) <= depth {
		counter = append(counter, 0)
	}
	counter[depth]++
	return counter
}

func recurse(auds int, feds int, limit int) {

	_, leaders, msgs := newElections(feds, auds, false)
	var msgpath []*DirectedMessage
	Dive(msgs, leaders, 0, limit, msgpath)
}

// reuse encoder/decoder so we don't recompile the struct definition
var enc *gob.Encoder
var dec *gob.Decoder

// LoopingDetected will the number of looping leaders
func LoopingDetected(global *election.Display) int {
	return global.DetectLoops()
}

func init() {
	buff := new(bytes.Buffer)
	enc = gob.NewEncoder(buff)
	dec = gob.NewDecoder(buff)
	MirrorMap.Init("dive")
}

func CloneElection(src *election.Election) *election.Election {
	return src.Copy()
	dst := new(election.Election)
	err := enc.Encode(src)
	if err != nil {
		errCollision++
	}
	err = dec.Decode(dst)
	if err != nil {
		errCollision++
	}
	return dst
}

// Create a Sha256 Hash from a byte array
func Sha(p []byte) [32]byte {
	b := sha256.Sum256(p)
	return b
}

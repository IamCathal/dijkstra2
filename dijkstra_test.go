package dijkstra

import (
	"log"
	"os"
	"reflect"
	"strconv"
	"testing"
)

//pq "github.com/Professorq/dijkstra"

func TestNoPath(t *testing.T) {
	testSolution(t, BestPath{}, ErrNoPath, "testdata/I.txt", 0, 4, true, -1)
}

func TestLoop(t *testing.T) {
	testSolution(t, BestPath{}, newErrLoop(2, 1), "testdata/J.txt", 0, 4, true, -1)
}

func TestCorrect(t *testing.T) {
	testSolution(t, getBSol(), nil, "testdata/B.txt", 0, 5, true, -1)
	testSolution(t, getKSolLong(), nil, "testdata/K.txt", 0, 4, false, -1)
	testSolution(t, getKSolShort(), nil, "testdata/K.txt", 0, 4, true, -1)
}

func TestCorrectSolutionsAll(t *testing.T) {
	graph := NewGraph()
	//Add the 3 verticies
	graph.AddVertex(0)
	graph.AddVertex(1)
	graph.AddVertex(2)
	graph.AddVertex(3)

	//Add the Arcs
	graph.AddArc(0, 1, 1)
	graph.AddArc(0, 2, 1)
	graph.AddArc(1, 3, 0)
	graph.AddArc(2, 3, 0)
	testGraphSolutionAll(t, BestPaths{BestPath{1, []int{0, 2, 3}}, BestPath{1, []int{0, 1, 3}}}, nil, *graph, 0, 3, true)
}

func TestCorrectAllLists(t *testing.T) {
	for i := 0; i <= 3; i++ {
		testSolution(t, getBSol(), nil, "testdata/B.txt", 0, 5, true, i)
		testSolution(t, getKSolLong(), nil, "testdata/K.txt", 0, 4, false, i)
		testSolution(t, getKSolShort(), nil, "testdata/K.txt", 0, 4, true, i)
	}
}

func TestCorrectAutoLargeList(t *testing.T) {
	g := NewGraph()
	for i := 0; i < 2000; i++ {
		v := g.AddNewVertex()
		v.AddArc(i+1, 1)
	}
	g.AddNewVertex()
	_, err := g.Shortest(0, 2000)
	testErrors(t, nil, err, "manual test")
	_, err = g.Longest(0, 2000)
	testErrors(t, nil, err, "manual test")
}

func BenchmarkSetup(b *testing.B) {
	nodeIterations := 6
	nodes := 1
	for j := 0; j < nodeIterations; j++ {
		nodes *= 4
		b.Run("setup/"+strconv.Itoa(nodes)+"Nodes", func(b *testing.B) {
			filename := "testdata/bench/" + strconv.Itoa(nodes) + ".txt"
			if _, err := os.Stat(filename); err != nil {
				g := Generate(nodes)
				err := g.ExportToFile(filename)
				if err != nil {
					log.Fatal(err)
				}
			}
			g, _ := Import(filename)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				g.setup(true, 0, -1)
			}
		})
	}
}

func benchmarkList(b *testing.B, nodes, list int, shortest bool) {

	filename := "testdata/bench/" + strconv.Itoa(nodes) + ".txt"
	if _, err := os.Stat(filename); err != nil {
		g := Generate(nodes)
		err := g.ExportToFile(filename)
		if err != nil {
			log.Fatal(err)
		}
	}
	graph, _ := Import(filename)
	src, dest := 0, len(graph.Verticies)-1
	//====RESET TIMER BEFORE LOOP====
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		graph.setup(shortest, src, list)
		graph.postSetupEvaluate(src, dest, shortest)
	}
}

func testSolution(t *testing.T, best BestPath, wanterr error, filename string, from, to int, shortest bool, list int) {
	var err error
	var graph Graph
	graph, err = Import(filename)
	if err != nil {
		t.Fatal(err, filename)
	}
	var got BestPath
	var gotAll BestPaths
	if list >= 0 {
		graph.setup(shortest, from, list)
		got, err = graph.postSetupEvaluate(from, to, shortest)
	} else if shortest {
		got, err = graph.Shortest(from, to)
	} else {
		got, err = graph.Longest(from, to)
	}
	testErrors(t, wanterr, err, filename)
	testResults(t, got, best, shortest, filename)
	if list >= 0 {
		graph.setup(shortest, from, list)
		gotAll, err = graph.postSetupEvaluateAll(from, to, shortest)
	} else if shortest {
		gotAll, err = graph.ShortestAll(from, to)
	} else {
		gotAll, err = graph.LongestAll(from, to)
	}
	testErrors(t, wanterr, err, filename)
	if len(gotAll) == 0 {
		gotAll = BestPaths{BestPath{}}
	}
	testResults(t, gotAll[0], best, shortest, filename)
}

func testGraphSolutionAll(t *testing.T, best BestPaths, wanterr error, graph Graph, from, to int, shortest bool) {
	var err error
	var gotAll BestPaths
	if shortest {
		gotAll, err = graph.ShortestAll(from, to)
	} else {
		gotAll, err = graph.LongestAll(from, to)
	}
	testErrors(t, wanterr, err, "From graph")
	if len(gotAll) == 0 {
		gotAll = BestPaths{BestPath{}}
	}
	testResultsGraphAll(t, gotAll, best, shortest)
}

func testResultsGraphAll(t *testing.T, got, best BestPaths, shortest bool) {
	distmethod := "Shortest"
	if !shortest {
		distmethod = "Longest"
	}
	if len(got) != len(best) {
		t.Error(distmethod, " amount of solutions incorrect\ngot: ", len(got), "\nwant: ", len(best))
		return
	}
	for i := range got {
		if got[i].Distance != best[i].Distance {
			t.Error(distmethod, " distance incorrect\ngot: ", got[i].Distance, "\nwant: ", best[i].Distance)
		}
	}
	for i := range got {
		found := false
		j := -1
		for j = range best {
			if reflect.DeepEqual(got[i].Path, best[j].Path) {
				//delete found result
				best = append(best[:j], best[j+1:]...)
				found = true
				break
			}
		}
		if found == false {
			t.Error(distmethod, " could not find path in solution\ngot:", got[i].Path)
		}
	}
}

func testResults(t *testing.T, got, best BestPath, shortest bool, filename string) {
	distmethod := "Shortest"
	if !shortest {
		distmethod = "Longest"
	}
	if got.Distance != best.Distance {
		t.Error(distmethod, " distance incorrect\n", filename, "\ngot: ", got.Distance, "\nwant: ", best.Distance)
	}
	if !reflect.DeepEqual(got.Path, best.Path) {
		t.Error(distmethod, " path incorrect\n\n", filename, "got: ", got.Path, "\nwant: ", best.Path)
	}
}

func getKSolLong() BestPath {
	return BestPath{
		31,
		[]int{
			0, 1, 3, 2, 4,
		},
	}
}
func getKSolShort() BestPath {
	return BestPath{
		2,
		[]int{
			0, 3, 4,
		},
	}
}

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"strconv"
	"strings"
)

type frac struct {
	n, d big.Int
}

func (f1 frac) sum(f2 frac) frac {
	var newN1, newN2, newD, newN big.Int

	newN1.Mul(&f1.n, &f2.d)
	newN2.Mul(&f1.d, &f2.n)
	newD.Mul(&f1.d, &f2.d)
	newN.Add(&newN1, &newN2)

	f := frac{n: newN, d: newD}
	f.simplify()

	return f
}

func (f1 frac) minus(f2 frac) frac {
	var newN1, newN2, newD, newN big.Int

	newN1.Mul(&f1.n, &f2.d)
	newN2.Mul(&f1.d, &f2.n)
	newD.Mul(&f1.d, &f2.d)
	newN.Sub(&newN1, &newN2)

	f := frac{n: newN, d: newD}
	f.simplify()

	return f
}

func (f1 frac) mul(f2 frac) frac {
	var mn, md big.Int
	mn.Mul(&f1.n, &f2.n)
	md.Mul(&f1.d, &f2.d)
	m := frac{n: mn, d: md}
	m.simplify()

	return m
}

func (f1 frac) invert() frac {
	return frac{n: f1.d, d: f1.n}
}

func (f1 frac) div(f2 frac) frac {
	return f1.mul(f2.invert())
}

func lcm(x, y big.Int) big.Int {
	var z big.Int
	z.Mul(z.Div(&x, z.GCD(nil, nil, &x, &y)), &y)
	return z
}

func (f1 *frac) simplify() {
	var z, one, zero big.Int
	one.SetInt64(1)
	zero.SetInt64(0)

	if f1.n.Cmp(&zero) == 0 {
		return
	}

	var absN, absD big.Int
	absN.Abs(&f1.n)
	absD.Abs(&f1.d)

	z.GCD(nil, nil, &absN, &absD)
	if z.Cmp(&one) == 0 {
		return
	}
	f1.n.Div(&f1.n, &z)
	f1.d.Div(&f1.d, &z)
	f1.simplify()
}

func (f1 frac) abs() frac {
	var absN big.Int
	absN.Abs(&f1.n)
	return frac{n: absN, d: f1.d}
}

func (f1 frac) neg() frac {
	return f1.mul(frac{n: *big.NewInt(-1), d: *big.NewInt(1)})
}

func (f1 frac) igte(i int64) bool {
	var newN big.Int
	newN.Mul(&f1.d, big.NewInt(i))
	f2 := frac{n: newN, d: f1.d}
	return f1.n.Cmp(&f2.n) >= 0
}

func (f1 frac) ilte(i int64) bool {
	var newN big.Int
	newN.Mul(&f1.d, big.NewInt(i))
	f2 := frac{n: newN, d: f1.d}
	return f1.n.Cmp(&f2.n) <= 0
}

func (f1 frac) ilt(i int64) bool {
	var newN big.Int
	newN.Mul(&f1.d, big.NewInt(i))
	f2 := frac{n: newN, d: f1.d}
	return f1.n.Cmp(&f2.n) < 0
}

func (f1 frac) eq(f2 frac) bool {
	return f1.n.Cmp(&f2.n) == 0 && f1.d.Cmp(&f2.d) == 0
}

type vertex struct {
	x, y frac
}

func (v1 vertex) eq(v2 vertex) bool { // vertex is simplified upon parse
	return v1.x.eq(v2.x) && v1.y.eq(v2.y)
}

type edge struct {
	v1, v2 vertex
}

func (e1 edge) isLengthZero() bool {
	return e1.v1.eq(e1.v2)
}

func (e1 edge) touchesOrCrosses(e2 edge) bool {
	return isVertexOnEdge(e1, e2.v1) || isVertexOnEdge(e1, e2.v2) || (isVertexRightOfEdge(e1, e2.v1) && isVertexRightOfEdge(e1, e2.v2))
}

func isVertexOnEdge(e edge, v vertex) bool { // but not on endpoints
	if v.eq(e.v1) || v.eq(e.v2) {
		return false
	}

	cp := crossProduct(vertex{x: e.v2.x.minus(e.v1.x), y: e.v2.y.minus(e.v1.y)}, vertex{x: v.x.minus(e.v1.x), y: v.y.minus(e.v1.y)})
	return cp.abs().eq(frac{n: *big.NewInt(0), d: *big.NewInt(1)})
}

func isVertexRightOfEdge(e edge, v vertex) bool {
	return crossProduct(vertex{x: e.v2.x.minus(e.v1.x), y: e.v2.y.minus(e.v1.y)}, vertex{x: v.x.minus(e.v1.x), y: v.y.minus(e.v1.y)}).ilt(0)
}

func crossProduct(v1 vertex, v2 vertex) frac {
	return v1.x.mul(v2.y).minus(v2.x.mul(v1.y))
}

type triangle struct {
	v1, v2, v3 vertex
}

type polygon struct {
	vertices []vertex
}

func (p polygon) edges() []edge {
	var es []edge
	for i := range p.vertices {
		if i < len(p.vertices) {
			es = append(es, edge{v1: p.vertices[i], v2: p.vertices[i+1]})
		}
	}
	return es
}

func (p polygon) hasLengthZeroEdges() bool {
	for _, e := range p.edges() {
		if e.isLengthZero() {
			return true
		}
	}
	return false
}

type problem struct {
	posPolys []polygon
	negPolys []polygon
	skeleton []edge
}

type solution struct {
	source      []vertex
	facets      []polygon
	destination []vertex
}

func parseProblem(s []string) problem {
	var p problem

	i := 0
	numPolys, _ := strconv.Atoi(s[i])
	for j := 0; j < numPolys; j++ {
		i++
		numVertices, _ := strconv.Atoi(s[i])
		i++
		poly := parsePoly(s[i : i+numVertices])
		if poly.isPositivePolygon() {
			p.posPolys = append(p.posPolys, poly)
		} else {
			p.negPolys = append(p.negPolys, poly)
		}
		i += numVertices
	}
	numEdges, _ := strconv.Atoi(s[i])
	i++
	p.skeleton = parseEdges(s[i : i+numEdges])

	return p
}

func parseSolution(s []string) solution {
	var sol solution

	var vs, vd []vertex
	var fs []polygon

	i := 0
	numVertices, _ := strconv.Atoi(s[i])
	i++
	for j := 0; j < numVertices; j++ {
		vs = append(vs, parseVertex(s[i]))
		i++
	}

	numFacets, _ := strconv.Atoi(s[i])
	i++
	for j := 0; j < numFacets; j++ {
		vertexIndices := strings.Split(s[i], " ")
		var pg polygon
		var vl []vertex
		for _, v := range vertexIndices {
			vi, _ := strconv.Atoi(v)
			vl = append(vl, vs[vi])
		}
		pg.vertices = vl
		fs = append(fs, pg)
		i++
	}

	for j := 0; j < numVertices; j++ {
		vd = append(vd, parseVertex(s[i]))
		i++
	}

	sol.source = vs
	sol.destination = vd

	return sol
}

func parsePoly(s []string) polygon {
	var vs []vertex
	for _, v := range s {
		vs = append(vs, parseVertex(v))
	}
	return polygon{vertices: vs}
}

func parseEdges(s []string) []edge {
	var es []edge
	for _, v := range s {
		vertices := strings.Split(v, " ")
		es = append(es, edge{v1: parseVertex(vertices[0]), v2: parseVertex(vertices[1])})
	}

	return es
}

// Simplifies vertices as it parses them (eases up equality)
func parseVertex(s string) vertex {
	fracs := strings.Split(s, ",")
	var bn1, bd1, bn2, bd2 big.Int
	f1 := strings.Split(fracs[0], "/")
	f2 := strings.Split(fracs[1], "/")
	n1 := f1[0]
	n2 := f2[0]
	var d1, d2 string
	if len(f1) > 1 {
		d1 = f1[1]
	} else {
		d1 = "1"
	}
	if len(f2) > 1 {
		d2 = f2[1]
	} else {
		d2 = "1"
	}
	bn1.SetString(n1, 10)
	bd1.SetString(d1, 10)
	bn2.SetString(n2, 10)
	bd2.SetString(d2, 10)

	x := frac{n: bn1, d: bd1}
	y := frac{n: bn2, d: bd2}
	x.simplify()
	y.simplify()
	return vertex{x: x, y: y}
}

func (p polygon) isPositivePolygon() bool {
	var zero big.Int
	zero.SetInt64(0)

	var auxN, auxD big.Int
	auxN.SetUint64(0)
	auxD.SetUint64(1)

	fracSum := frac{n: auxN, d: auxD}

	vs := p.vertices
	for i := 0; i < len(vs)-1; i++ {
		fracSum = fracSum.sum(vs[i+1].x.sum(vs[i].x).mul(vs[i+1].y.minus(vs[i].y)))
	}
	return fracSum.n.Cmp(&zero) >= 0
}

func (p polygon) triangles() []triangle {
	if len(p.vertices) < 3 {
		return []triangle{}
	}

	var ts []triangle
	for i := 2; i < len(p.vertices); i++ {
		ts = append(ts, triangle{v1: p.vertices[0], v2: p.vertices[i-1], v3: p.vertices[i]})
	}

	return ts
}

func (t triangle) area() frac {
	t.v2.x.minus(t.v1.x)

	xa := t.v2.x.minus(t.v1.x)
	ya := t.v2.y.minus(t.v1.y)
	xb := t.v3.x.minus(t.v1.x)
	yb := t.v3.y.minus(t.v1.y)

	m1 := xa.mul(yb)
	m2 := xb.mul(ya)

	return m1.minus(m2).abs().mul(frac{n: *big.NewInt(1), d: *big.NewInt(2)})
}

func (p polygon) area() frac {
	total := frac{n: *big.NewInt(0), d: *big.NewInt(1)}
	for _, t := range p.triangles() {
		total = total.sum(t.area())
	}
	return total
}

func (p problem) area() frac {
	total := frac{n: *big.NewInt(0), d: *big.NewInt(1)}
	for _, poly := range p.posPolys {
		total = total.sum(poly.area())
	}
	for _, poly := range p.negPolys {
		total = total.minus(poly.area())
	}

	return total
}

func (s solution) isSourceWithinBoundaries() bool {
	result := true
	for _, v := range s.source {
		if !(v.x.igte(0) && v.x.ilte(1) && v.y.igte(0) && v.y.ilte(1)) {
			result = false
		}
	}
	return result
}

func (s solution) areSourceVerticesUnique() bool { // vertices are already simplified
	for i, v := range s.source {
		for j, w := range s.source {
			if i != j && v.eq(w) {
				return false
			}
		}
	}
	return true
}

func (s solution) hasLengthZeroEdgeFacets() bool {

	for _, f := range s.facets {
		if f.hasLengthZeroEdges() {
			return true
		}
	}
	return false
}

func (s solution) hasNoSourceEdgeCrossesOrTouches() bool {
	edges := s.sourceEdges()
	for i, e := range edges {
		for j, f := range edges {
			if i != j && e.touchesOrCrosses(f) {
				fmt.Printf("Unfortunately %+v touches or crosses %+v\n", e, f)
				return false
			}
		}
	}
	return true
}

func (s solution) sourceEdges() []edge {
	var es []edge
	for i := range s.source {
		if i > 0 && i < len(s.source) {
			es = append(es, edge{v1: s.source[i-1], v2: s.source[i]})
		}
	}
	return es
}

func (s solution) validate() bool {
	return s.isSourceWithinBoundaries() && s.areSourceVerticesUnique() && !s.hasLengthZeroEdgeFacets() && s.hasNoSourceEdgeCrossesOrTouches()
}

func main() {
	var problemFile, solutionFile string

	flag.StringVar(&solutionFile, "solution", "", "Parse solution in specified file.")
	flag.StringVar(&problemFile, "problem", "", "Parse problem in specified file.")
	flag.Parse()

	if problemFile != "" {
		bProblemString, err := ioutil.ReadFile(problemFile)
		if err != nil {
			log.Fatalf("Could not read source problem file [%v]. err=%v", problemFile, err)
		}
		problem := parseProblem(strings.Split(string(bProblemString), "\n"))
		fmt.Printf("Problem:\n\n%+v\n\n", problem)
		fmt.Printf("Problem area: %v\n\n\n`", problem.area())
	}

	if solutionFile != "" {
		bSolutionString, err := ioutil.ReadFile(solutionFile)
		if err != nil {
			log.Fatalf("Could not read source solution file [%v]. err=%v", problemFile, err)
		}
		solution := parseSolution(strings.Split(string(bSolutionString), "\n"))
		fmt.Printf("Solution:\n\n%+v\n\n", solution)
	}
}

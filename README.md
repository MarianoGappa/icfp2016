# ICFP 2016
A valiant-ish attempt at ICFP 2016

[Log in](http://2016sv.icfpcontest.org/)

[Leaderboard](http://2016sv.icfpcontest.org/leaderboard)

[Blog, rules, news](http://icfpc2016.blogspot.com/)

## Made as part of the Movio team 

[Full Movio Repo](https://bitbucket.org/moviohq/icfp2016/)

## Running (requires Go)

```
$ go build .
./icfp2016 --problem test_problem.txt --solution test_solution.txt
```
(both arguments are optional, but if you put nothing it does nothing :P)

## Features

- Parses problem
- Calculates problem area, distinguishing positive and negative polygons
- Parses solution
- Partially validates solution
- All fractional calculations maintain full precision

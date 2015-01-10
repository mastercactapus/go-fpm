package main

import (
	"io"
	"sort"
)

type DependencyTree struct {
	Nodes map[string]DependencyNode
}
type DependencyNode struct {
	Name    string
	Version string
	Tarball string
	Shasum  string
	Nodes   map[string]DependencyNode
}

func (n *DependencyTree) Print(w io.Writer) {
	io.WriteString(w, ".\n")
	keys := sortedDepKeys(n.Nodes)
	i := 0
	for _, k := range keys {
		i++
		v := n.Nodes[k]
		v.Print("", i == len(keys), w)
	}
}
func (n *DependencyNode) Print(prefix string, last bool, w io.Writer) {
	if last {
		io.WriteString(w, prefix+"└── "+n.Name+"@"+n.Version+"\n")
		prefix = prefix + "    "
	} else {
		io.WriteString(w, prefix+"├── "+n.Name+"@"+n.Version+"\n")
		prefix = prefix + "│   "
	}
	keys := sortedDepKeys(n.Nodes)
	i := 0
	for _, k := range keys {
		i++
		v := n.Nodes[k]
		v.Print(prefix, i == len(keys), w)
	}
}
func sortedDepKeys(m map[string]DependencyNode) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func CalculateTree(r *Registry, deps DependencyMap) (*DependencyTree, error) {
	t := new(DependencyTree)
	r.cacheAll(deps)
	for name, req := range deps {
		vers, err := r.LatestCompatablePackageVersion(name, req)
		if err != nil {
			return nil, err
		}
		pkg, err := r.PackageByVersion(name, vers.String())
		if err != nil {
			return nil, err
		}
		t.Nodes[name] = DependencyNode{name, vers.String(), pkg.Dist.Tarball, pkg.Dist.Shasum, map[string]DependencyNode{}}
	}
	return t, nil
}

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/blang/semver"
	"io/ioutil"
	"net/http"
	"sort"
	"sync"
)

type Registry struct {
	baseURL    string
	cache      packageCache
	fetchQueue chan packageDataRequest
}

type packageCache map[string]*repoPackageData
type packageDataRequest struct {
	ch   chan *repoPackageData
	name string
}

type repoPackageData struct {
	Tags           map[string]string `json:"dist-tags"`
	Versions       map[string]*Package
	Name           string
	sortedVersions semver.Versions
}

type Package struct {
	Name                 string
	Version              string
	Dependencies         DependencyMap
	DevDependencies      DependencyMap
	OptionalDependencies DependencyMap
	PeerDependencies     DependencyMap
	Dist                 struct {
		Tarball string
		Shasum  string
	}
}

type DependencyMap map[string]*SemverRequirements

type ResponseError struct {
	Code   int
	Status string
}

type SatisfiesChecker interface {
	SatisfiedBy(semver.Version) bool
	String() string
}

func (r *ResponseError) Error() string {
	return fmt.Sprintf("Bad or unexpected status code %i: %s", r.Code, r.Status)
}

func (d *DependencyMap) UnmarshalJSON(data []byte) error {
	m := make(map[string]string, 100)
	err := json.Unmarshal(data, &m)
	if err != nil {
		log.Debugf("Failed to unmarshal dependencies as map[string]string: %s\n", err.Error())
		return err
	}

	deps := make(DependencyMap, len(m))
	for k, v := range m {
		req, err := NewSemverRequirements(v)
		if err != nil {
			log.Debugf("Failed to parse semver requirement '%s': %s\n", v, err.Error())
			return err
		}
		deps[k] = req
	}
	*d = deps
	return nil
}

func NewRegistry(baseUrl string) *Registry {
	r := new(Registry)
	//Ensure trailing '/'
	if baseUrl[len(baseUrl)-1:] == "/" {
		r.baseURL = baseUrl
	} else {
		r.baseURL = baseUrl + "/"
	}
	r.cache = make(packageCache, 200)
	r.fetchQueue = make(chan packageDataRequest, 200)
	go r.dataFetchLoop()
	return r
}

func (r *Registry) CompatablePackageVersions(name string, req SatisfiesChecker) ([]semver.Version, error) {
	versions, err := r.PackageVersions(name)
	if err != nil {
		return nil, err
	}
	result := make([]semver.Version, 0, len(versions))
	for _, v := range versions {
		if req.SatisfiedBy(v) {
			result = append(result, v)
		}
	}
	return result, nil
}

func (r *Registry) LatestPackageVersion(name string) (v semver.Version, err error) {
	versions, err := r.PackageVersions(name)
	if err != nil {
		return v, err
	}
	if len(versions) == 0 {
		return v, errors.New("No versions available for: " + name)
	}
	return versions[0], nil
}

func (r *Registry) LatestCompatablePackageVersion(name string, req SatisfiesChecker) (version semver.Version, err error) {
	versions, err := r.PackageVersions(name)
	if err != nil {
		return version, err
	}
	for _, v := range versions {
		if req.SatisfiedBy(v) {
			return v, nil
		}
	}
	return version, errors.New("No compatable versions available for: " + name + "@" + req.String())
}

func (r *Registry) PackageByVersion(name string, version string) (*Package, error) {
	p, err := r.packageData(name)
	if err != nil {
		return nil, err
	}
	if p.Versions[version] == nil {
		return nil, errors.New("No version found for: " + name + "@" + version)
	}
	return p.Versions[version], nil
}

func (r *Registry) PackageVersions(name string) (semver.Versions, error) {
	p, err := r.packageData(name)
	if err != nil {
		return nil, err
	}
	return p.sortedVersions, nil
}

func (r *Registry) dataFetchLoop() {
	log.Debugln("Started loop")
	pending := make(map[string][]chan *repoPackageData, 100)
	complete := make(chan *repoPackageData, 100)
	for {
		select {
		case req := <-r.fetchQueue:
			log.Debugln("Processing", req.name, "from queue.")
			//if already cached and good to go then return
			if r.cache[req.name] != nil {
				if req.ch == nil {
					continue
				}
				req.ch <- r.cache[req.name]
				//if already being fetch then add the return channel
			} else if pending[req.name] != nil {
				if req.ch == nil {
					continue
				}
				pending[req.name] = append(pending[req.name], req.ch)
				//initiate a new fetch
			} else {
				pending[req.name] = make([]chan *repoPackageData, 0, 20)
				if req.ch != nil {
					pending[req.name] = append(pending[req.name], req.ch)
				}
				go func(name string) {
					data, err := r.fetchPackageData(req.name)
					if err != nil {
						log.Fatalln(err)
					}
					complete <- data
				}(req.name)
			}
		case data := <-complete:
			log.Debugln("Completed", data.Name)
			r.cache[data.Name] = data
			if pending[data.Name] != nil {
				for _, v := range pending[data.Name] {
					v <- data
				}
				pending[data.Name] = nil
			}
		}
	}
}

func (r *Registry) cacheAll(deps DependencyMap) {
	var wg sync.WaitGroup
	for k := range deps {
		wg.Add(1)
		go func(name string) {
			r.packageData(k)
			wg.Done()
		}(k)
	}
	wg.Wait()
}

func (r *Registry) packageData(name string) (*repoPackageData, error) {
	ch := make(chan *repoPackageData, 1)
	r.fetchQueue <- packageDataRequest{ch, name}
	return <-ch, nil
}

func (r *Registry) fetchPackageData(name string) (*repoPackageData, error) {
	fullUrl := r.baseURL + name
	log.Debugf("Fetch data for '%s' from: %s", name, fullUrl)
	res, err := http.Get(fullUrl)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		ioutil.ReadAll(res.Body)
		return nil, &ResponseError{res.StatusCode, res.Status}
	}
	p := new(repoPackageData)
	d := json.NewDecoder(res.Body)
	err = d.Decode(p)
	if err != nil {
		return nil, err
	}
	p.sortedVersions = make(semver.Versions, 0, len(p.Versions))
	compareLatest := true
	latest, err := semver.New(p.Tags["latest"])
	if err != nil {
		compareLatest = false
	}
	for k := range p.Versions {
		sv, err := semver.New(k)
		if err != nil {
			return nil, err
		}
		if !compareLatest || sv.LTE(latest) {
			p.sortedVersions = append(p.sortedVersions, sv)
		}
	}
	sort.Sort(sort.Reverse(p.sortedVersions))
	return p, nil
}

package fakesysfs

import "fmt"

type FakeSysFs struct {
	distances    map[int]string
	distancesErr error
}

func (fs *FakeSysFs) GetDistances(nodeId int) (string, error) {
	if fs.distancesErr != nil {
		return "", fs.distancesErr
	}

	if _, ok := fs.distances[nodeId]; !ok {
		return "", fmt.Errorf("distance not found")
	}

	return fs.distances[nodeId], nil
}

func (fs *FakeSysFs) SetDistances(nodeId int, distances string, err error) {
	if fs.distances == nil {
		fs.distances = map[int]string{nodeId: distances}
	} else {
		fs.distances[nodeId] = distances
	}
	fs.distancesErr = err
}

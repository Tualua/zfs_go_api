package main

import (
	"fmt"
	"log"
	"strconv"

	zfs "github.com/bicomsystems/go-libzfs"
)

/*
type ZfsEntity struct {
	XMLName    xml.Name `xml:"zfsentity"`
	Name       string   `xml:"name"`
	Used       string   `xml:"used"`
	Avail      string   `xml:"avail"`
	Refer      string   `xml:"refer"`
	MountPoint string   `xml:"mountpoint"`
}
*/
type ZfsEntity struct {
	Name       string `json:"name"`
	Used       string `json:"used"`
	Avail      string `json:"avail"`
	Refer      string `json:"refer"`
	MountPoint string `json:"mountpoint"`
}

func zfsGetProperties(ds *zfs.Dataset) ZfsEntity {
	var (
		res ZfsEntity
	)
	if prop, err := ds.GetProperty(zfs.DatasetPropName); err != nil {
		fmt.Println(err)
	} else {
		res.Name = prop.Value
	}
	if prop, err := ds.GetProperty(zfs.DatasetPropUsed); err != nil {
		fmt.Println(err)
	} else {
		res.Used = prop.Value
	}
	if prop, err := ds.GetProperty(zfs.DatasetPropAvailable); err != nil {
		res.Avail = "-"
	} else {
		res.Avail = prop.Value
	}
	if prop, err := ds.GetProperty(zfs.DatasetPropReferenced); err != nil {
		fmt.Println(err)
	} else {
		res.Refer = prop.Value
	}
	if prop, err := ds.GetProperty(zfs.DatasetPropMountpoint); err != nil {
		res.MountPoint = "-"
	} else {
		res.MountPoint = prop.Value
	}
	return res
}

func zfsGetChildren(ds *zfs.Dataset) []*zfs.Dataset {
	var (
		res []*zfs.Dataset
	)
	if len(ds.Children) == 0 {
		return append(res, ds)
	} else {
		res = append(res, ds)
		for _, v := range ds.Children {
			res = append(res, zfsGetChildren(&v)...)
		}
	}

	return res
}

func ZfsListAll() ([]ZfsEntity, error) {
	var (
		ds  []*zfs.Dataset
		res []ZfsEntity
		err error
	)
	if datasets, err := zfs.DatasetOpenAll(); err != nil {
		log.Println(err.Error())
	} else {
		for _, v := range datasets {
			ds = append(ds, zfsGetChildren(&v)...)
		}
		for _, v := range ds {
			res = append(res, zfsGetProperties(v))
		}
		zfs.DatasetCloseAll(datasets)
	}
	return res, err
}

func ZfsCreateSnapshot(snapsource string, snapname string) error {
	var (
		err error
		rd  zfs.Dataset
	)
	props := make(map[zfs.Prop]zfs.Property)

	if rd, err = zfs.DatasetSnapshot(fmt.Sprintf("%s@%s", snapsource, snapname), false, props); err != nil {
		log.Println(err.Error())
	} else {
		path, _ := rd.Path()
		log.Printf("Snapshot %s created\n", path)
	}
	return err
}

func ZfsGetLastSnapshot(DsPath string) (string, error) {
	var (
		res   string
		err   error
		maxTs int64 = 0
		ds    zfs.Dataset
	)

	if ds, err = zfs.DatasetOpen(DsPath); err != nil {
		log.Println(err.Error())
	} else {
		if dsSnapshots, err := ds.Snapshots(); err != nil {
			log.Println(err.Error())
		} else {
			for _, s := range dsSnapshots {
				path, _ := s.Path()
				creation, _ := s.GetProperty(zfs.DatasetPropCreation)
				ts, _ := strconv.ParseInt(creation.Value, 10, 64)
				if ts >= maxTs {
					maxTs = ts
					res = path
				}
			}
		}
	}

	return res, err
}

func ZfsGetCloneInfo(ClonePath string) (map[string]string, error) {
	var (
		res map[string]string
		err error
		ds  zfs.Dataset
	)
	res = make(map[string]string)
	if ds, err = zfs.DatasetOpenSingle(ClonePath); err != nil {
		log.Println(err.Error())
	} else {
		propOrigin, _ := ds.GetProperty(zfs.DatasetPropOrigin)
		res["origin"] = propOrigin.Value
		propWritten, _ := ds.GetProperty(zfs.DatasetPropWritten)
		res["written"] = propWritten.Value
	}
	return res, err
}

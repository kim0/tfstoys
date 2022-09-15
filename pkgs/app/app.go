package app

import (
	"fmt"
	"io"
	"log"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dustin/go-humanize"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/kim0/tfstoys/pkgs/remotestate"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

type Since_Strategy struct {
	Days uint
}

func Diff(state_bucket string, state_path string, ss Since_Strategy) {
	var idx int
	var err error

	logrus.Info("Listing state files from remote S3 backend: ", state_bucket)

	var maxkeys int64 = 1000
	objects_resp := remotestate.ListBucketObjects(state_bucket, "", &maxkeys)
	objects := objects_resp.Contents
	sort.SliceStable(objects, func(i int, j int) bool {
		return (*objects[i].LastModified).Unix() > (*objects[j].LastModified).Unix()
	})

	if state_path != "" {
		idx = slices.IndexFunc(objects, func(o *s3.Object) bool {
			return *o.Key == state_path || (fmt.Sprintf("%s/terraform.tfstate", state_path) == *o.Key)
		})
	}

	if state_path == "" || idx == -1 {
		logrus.Warn("Exact match not found. Attmepting fuzzyfind")
		idx, err = fuzzyfinder.Find(
			objects,
			func(i int) string {
				return fmt.Sprintf("%s (%s) [%s]", *objects[i].Key, objects[i].LastModified, humanize.Bytes(uint64(*objects[i].Size)))
			},
			fuzzyfinder.WithHeader("Choose state file. Fuzzy finding (do not use space)"),
		)
		if err != nil {
			log.Fatal(err)
		}
	}
	logrus.Info(fmt.Sprintf("State selected: %s\n", *objects[idx].Key))
	state := objects[idx]

	versions := remotestate.GetObjectVersions(state_bucket, *state.Key)

	var compare_idx []int
	if ss.Days != 0 {
		for i, v := range versions[:len(versions)-1] {
			vplus := versions[i+1]
			vl := *v.LastModified
			vpl := *vplus.LastModified
			if vl.Sub(vpl) > (time.Duration(ss.Days) * 24 * time.Hour) {
				compare_idx = append(compare_idx, i+1)
				break
			}
		}
		compare_idx = append(compare_idx, 0)
		logrus.Info("Comparing the two versions: %#v %#v\n", compare_idx[0], compare_idx[1])
		if len(compare_idx) != 2 {
			log.Fatal("Cannot find two versions with the requested gap!")
		}
	} else {
		compare_idx, err = fuzzyfinder.FindMulti(
			versions,
			func(i int) string {
				return fmt.Sprintf("%s (%s) [%s]", *versions[i].Key, humanize.Time(*versions[i].LastModified), humanize.Bytes(uint64(*versions[i].Size)))
			},
			fuzzyfinder.WithHeader("Choose versions to companre (use TAB key)"),
		)
		if err != nil {
			log.Fatal(err)
		}
	}
	if len(compare_idx) != 2 {
		logrus.Fatal("Must select only 2 versions to compare! Cannot continue")
	}

	logrus.Debug("Versions to compare:", compare_idx)
	v1 := remotestate.GetBucketObjects(state_bucket, *objects[idx].Key, versions[compare_idx[0]])
	defer v1.Body.Close()
	v1_bytes, err := io.ReadAll(v1.Body)
	if err != nil {
		log.Fatal(err)
	}
	v1_content := string(v1_bytes)
	logrus.Info("V1 read bytes: ", len(v1_content))

	v2 := remotestate.GetBucketObjects(state_bucket, *objects[idx].Key, versions[compare_idx[1]])
	defer v2.Body.Close()
	v2_bytes, err := io.ReadAll(v2.Body)
	if err != nil {
		log.Fatal(err)
	}
	v2_content := string(v2_bytes)
	logrus.Info("V2 read bytes: ", len(v2_content))

	v1_filename := fmt.Sprintf("%s.%s", *objects[idx].Key, *versions[compare_idx[0]].VersionId)
	v2_filename := fmt.Sprintf("%s.%s", *objects[idx].Key, *versions[compare_idx[1]].VersionId)
	edits := myers.ComputeEdits(span.URIFromPath(v1_filename), string(v1_content), string(v2_content))
	diff := fmt.Sprint(gotextdiff.ToUnified(v1_filename, v2_filename, string(v1_content), edits))

	// fmt.Println(len(diff))
	fmt.Println(diff)
}

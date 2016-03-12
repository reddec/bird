package gazer

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/reddec/bird"
)

func nest(params url.Values) (bird.Bird, error) {
	return func(kill <-chan int) error {
	LOOP:
		for {
			select {
			case <-kill:
				break LOOP
			case <-time.After(1 * time.Second):
				log.Println("Hello from " + params.Get("name"))
			}
		}
		return nil
	}, nil
}

func getBirds() ([]birdFace, error) {
	var faces []birdFace
	resp, err := http.Get("http://127.0.0.1:9090/birds")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	return faces, decoder.Decode(&faces)
}

func landBirds() error {
	req, err := http.NewRequest("PUT", "http://127.0.0.1:9090/birds?action=land", bytes.NewBufferString(""))
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		data, _ := ioutil.ReadAll(res.Body)
		return errors.New("bad code: " + string(data))
	}
	return nil
}

func createBirds(name string, interval time.Duration, raise bool) (birdFace, error) {
	var face birdFace
	values := url.Values{}
	values.Add("name", name)
	values.Add("interval", interval.String())
	if raise {
		values.Add("raise", "1")
	}
	resp, err := http.PostForm("http://127.0.0.1:9090/birds", values)
	if err != nil {
		return face, err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	return face, decoder.Decode(&face)
}

func TestGazer(t *testing.T) {

	gz := NewGazer(bird.NewFlock(), nest)
	http.Handle("/", gz)
	go http.ListenAndServe(":9090", nil)
	time.Sleep(1 * time.Second)

	birds, err := getBirds()
	if err != nil {
		t.Fatal("GET", err)
	}
	if len(birds) != 0 {
		t.Fatal("No birds expected")
	}
	// create
	brd, err := createBirds("wooo", 2*time.Second, true)
	if err != nil {
		t.Fatal("POST", err)
	}
	if brd.Name != "wooo" {
		t.Fatal("Created bad name")
	}
	if brd.Interval != 2*time.Second {
		t.Fatal("Created bad interval")
	}
	if !brd.Flying {
		t.Fatal("Not flying")
	}
	// Select again
	birds, err = getBirds()
	if err != nil {
		t.Fatal("GET", err)
	}
	if len(birds) != 1 {
		t.Fatal("1 bird expected")
	}

	{ //check again but as one item from list
		brd = birds[0]
		if brd.Name != "wooo" {
			t.Fatal("Created bad name")
		}
		if brd.Interval != 2*time.Second {
			t.Fatal("Created bad interval")
		}
		if !brd.Flying {
			t.Fatal("Not flying")
		}
	}
	// Land
	err = landBirds()
	if err != nil {
		t.Fatal("PUT", err)
	}

	birds, err = getBirds()
	if err != nil {
		t.Fatal("GET", err)
	}
	if len(birds) != 1 {
		t.Fatal("1 bird expected")
	}
	if birds[0].Flying {
		t.Fatal("Flying yet")
	}
}

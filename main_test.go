package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)

	// Discard log messages during normal testing
	logrus.SetOutput(ioutil.Discard)

	os.Exit(m.Run())
}

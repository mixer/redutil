package heartbeat_test

import (
	"errors"
	"testing"
	"time"

	"github.com/mixer/redutil/conn"
	"github.com/mixer/redutil/heartbeat"
	"github.com/mixer/redutil/test"
	"github.com/stretchr/testify/suite"
)

type SimpleDetectorSuite struct {
	*test.RedisSuite
}

func TestSimpleDetectorSuite(t *testing.T) {
	pool, _ := conn.New(conn.ConnectionParam{
		Address: "127.0.0.1:6379",
	}, 1)

	suite.Run(t, &SimpleDetectorSuite{test.NewSuite(pool)})
}

func (suite *SimpleDetectorSuite) TestConstruction() {
	d := heartbeat.NewDetector("foo", suite.Pool, heartbeat.HashExpireyStrategy{time.Second})

	suite.Assert().IsType(heartbeat.SimpleDetector{}, d)
}

func (suite *SimpleDetectorSuite) TestDetectDelegatesToStrategy() {
	strategy := &TestStrategy{}
	strategy.On("Expired", "foo", suite.Pool).Return([]string{}, nil)

	d := heartbeat.NewDetector("foo", suite.Pool, strategy)
	d.Detect()

	strategy.AssertCalled(suite.T(), "Expired", "foo", suite.Pool)
}

func (suite *SimpleDetectorSuite) TestDetectPropogatesValues() {
	strategy := &TestStrategy{}
	strategy.On("Expired", "foo", suite.Pool).Return([]string{"foo", "bar"}, errors.New("baz"))

	d := heartbeat.NewDetector("foo", suite.Pool, strategy)
	expired, err := d.Detect()

	suite.Assert().Equal(expired, []string{"foo", "bar"})
	suite.Assert().Equal(err.Error(), "baz")
}

func (suite *SimpleDetectorSuite) TestDetectPurgesData() {
	strategy := &TestStrategy{}
	strategy.On("Purge", "foo", "id1", suite.Pool).Return(nil).Once()
	strategy.On("Purge", "foo", "id2", suite.Pool).Return(errors.New("baz")).Once()

	d := heartbeat.NewDetector("foo", suite.Pool, strategy)
	suite.Assert().Nil(d.Purge("id1"))
	suite.Assert().Equal(d.Purge("id2").Error(), "baz")
}

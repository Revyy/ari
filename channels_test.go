package ari

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type ChannelTests struct {
	suite.Suite
	c    chan *Event
	list []Channel
}

func (s *ChannelTests) SetupSuite() {
	DefaultClient.Go()
	time.Sleep(1 * time.Second)
}

func (s *ChannelTests) TearDownSuite() {
	DefaultClient.Close()
}

func (s *ChannelTests) TestCreateChannelByDialplan() {
	req := OriginateRequest{
		Endpoint:  "PJSIP/102",
		Extension: "203",
		Context:   "default",
		Priority:  1,
	}
	channel, err := DefaultClient.CreateChannel(req)
	s.Nil(err, "Channel by Dialplan not made")

	// Try to get the channel back
	channel2, err := DefaultClient.GetChannel(channel.Id)
	s.Nil(err, "Could not access specific channel MyChannelId")
	s.Equal(channel, channel2, "Returned channels are not equal.")
	fmt.Println("Channel returned:", channel2)

	err = DefaultClient.HangupChannel(channel.Id, "")
	s.Nil(err, "Error hanging up channel created by dialplan")

}

func (s *ChannelTests) TestCreateChannelByApp() {
	req2 := OriginateRequest{
		Endpoint:  "PJSIP/101",
		App:       "default",
		AppArgs:   "",
		ChannelId: "MyApp",
	}

	Chan2, err := DefaultClient.CreateChannel(req2)
	s.Nil(err, "Channel by Application not made")
	s.NotEmpty(Chan2.Id)

	s.Equal("MyApp", Chan2.Id)

	//Wait until we receive "StasisStart" from our channel, meaning it has been answered (on far end) and may be answered by Asterisk.

	e := <-DefaultClient.Events
	for e.Type != "StasisStart" {
		e = <-DefaultClient.Events
	}

	err = DefaultClient.AnswerChannel("MyApp")
	s.Nil(err)
}

func (s *ChannelTests) TestListChannel() {
	list, err := DefaultClient.ListChannels()
	s.Nil(err, "Could not access channel lists", err)
	for _, element := range list {
		fmt.Println(element)
	}
}

func (s *ChannelTests) TestRingChannel() {
	err := DefaultClient.ChannelRing("MyApp")
	s.Nil(err, "Could not ring to channel 'MyApp'")
	fmt.Println("sleeping 1 second while we wait for a ring")
	time.Sleep(1 * time.Second)
	err = DefaultClient.StopRinging("MyApp")
	s.Nil(err, "Could not stop ringing to channel 'MyApp'")
}

func (s *ChannelTests) TestMuteChannel() {
	err := DefaultClient.MuteChannel("MyApp", "in")
	s.Nil(err, "Could not mute channel 'MyApp'")
	fmt.Println("Sleeping 1 second while we wait for mute")
	time.Sleep(1 * time.Second)
	err = DefaultClient.UnMuteChannel("MyApp", "out")
	s.Nil(err, "Could not stop mute on channel 'MyApp'")
}

func (s *ChannelTests) TestHoldChannel() {
	err := DefaultClient.HoldChannel("MyApp")
	s.Nil(err, "Could not hold channel.")
	fmt.Println("Sleeping 1 second while we wait for hold")
	time.Sleep(1 * time.Second)
	err = DefaultClient.StopHoldChannel("MyApp")
	s.Nil(err, "Could not un-hold channel.")
}

//FIXME This does not work.
func (s *ChannelTests) TestDTMFSend() {
	fmt.Println("Sending 1 to channel")
	req := SendDTMFToChannelRequest{
		Dtmf: "1",
	}
	err := DefaultClient.SendDTMFToChannel("MyApp", req)
	s.Nil(err, "Could not send DTMF to channel.")
}

func (s *ChannelTests) TestMoh() {
	err := DefaultClient.PlayMOHToChannel("MyApp", "default")
	s.Nil(err, "Could not play MOH to channel 'MyApp'")
	fmt.Println("Sleeping 1 second while we wait for moh")
	time.Sleep(1 * time.Second)
	err = DefaultClient.StopMohChannel("MyApp")
	s.Nil(err, "Could not stop MOH on channel 'MyApp'")
}

func (s *ChannelTests) TestSilence() {
	err := DefaultClient.PlaySilenceToChannel("MyApp")
	s.Nil(err, "Could not send silence to channel 'MyApp'")
	fmt.Println("Sleeping 1 second while we wait for silence")
	time.Sleep(1 * time.Second)
	err = DefaultClient.StopSilenceChannel("MyApp")
	s.Nil(err, "Could not stop silence on channel 'MyApp'")
}

func (s *ChannelTests) TestSetAndGetChannelVariable() {
	err := DefaultClient.SetChannelVariable("MyApp", "testVariable", "success")
	s.Nil(err, "Could not set 'testVariable' to 'success'")
	variable, err := DefaultClient.GetChannelVariable("MyApp", "testVariable")
	s.Nil(err, "error retrieving 'testVariable' value")
	s.Equal(variable.Value, "success", "'testVariable' not equal to 'success'")
}

func (s *ChannelTests) TestLiveRecording() {
	req := RecordRequest{
		Name:               "name",
		Format:             "wav",
		Beep:               true,
		IfExists:           "overwrite",
		MaxDurationSeconds: 4,
	}
	rec, err := DefaultClient.RecordChannel("MyApp", req)
	fmt.Println("Allowing 5 seconds for liveRecording")
	time.Sleep(5 * time.Second)
	s.Nil(err, "Could not start live recording 'name'")
	s.Equal(rec.Name, "name", "Name retrieved does not match for live recording")
	s.Equal(rec.Format, "wav", "Format retrieved does not match for live recording")

	req2 := PlayMediaRequest{
		Media: "recording:name",
	}
	_, err = DefaultClient.PlayToChannel("MyApp", req2)
	fmt.Println("Waiting 5 seconds to play back recording")
	time.Sleep(5 * time.Second)
	s.Nil(err, "Couldn't play recording back to 'MyApp'")

	fmt.Println("Deleting recording 'name' for future tests.")
	err = DefaultClient.DeleteStoredRecording("name")
	s.Nil(err, "Couldn't delete stored recording.")
}

//Should be performed last, so Z is thrown into name.
func (s *ChannelTests) TestZHangup() {
	err := DefaultClient.HangupChannel("MyApp", "")
	s.Nil(err, "Error hanging up channel 'MyApp'")
}

func (s *ChannelTests) TestYContinue() {

	req := ContinueChannelRequest{
		Extension: "600",
		Context:   "default",
		Priority:  1,
	}

	err := DefaultClient.ContinueChannel("MyApp", req)
	s.Nil(err, "Could not exit application -> dialplan.")
	fmt.Println("Sleeping 1 second to exit application for continue.")
	time.Sleep(1 * time.Second)
}

//Snooping requires an outside channel TODO.

func TestChannelSuite(t *testing.T) {
	suite.Run(t, new(ChannelTests))
}

package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	lru "github.com/hashicorp/golang-lru"
	"github.com/jpillora/backoff"
	pb "github.com/mik3y/flightmon/proto"
)

func safeParseInt(s string) (*int32, error) {
	if len(s) == 0 {
		return nil, nil
	}
	val, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return nil, err
	}
	retVal := int32(val)
	return &retVal, nil
}

func safeParseFloat(s string) (*float32, error) {
	if len(s) == 0 {
		return nil, nil
	}
	val, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return nil, err
	}
	retVal := float32(val)
	return &retVal, nil
}

func nilableString(s string) *string {
	if len(s) == 0 {
		return nil
	}
	trimmed := strings.TrimSpace(s)
	return &trimmed
}

func processMessage(message string, trackedCache *lru.Cache, updates chan<- *pb.PositionUpdate) error {
	msg, err := parseMessage(message)
	if err != nil {
		return err
	}

	changed := false

	ident := msg.GetHexIdent()
	record := &pb.PositionUpdate{
		IcaoId: &ident,
	}

	var existingRecord *pb.PositionUpdate

	val, exists := trackedCache.Peek(ident)
	if exists {
		existingRecord = val.(*pb.PositionUpdate)
	} else {
		logger.Infof("Tracking new aircraft: %v", ident)
		existingRecord = nil
		changed = true
	}

	if msg.Callsign != nil && (record.Callsign == nil || (*record.Callsign != *msg.Callsign)) {
		record.Callsign = msg.Callsign
		changed = true
	}

	if msg.Lat != nil && (record.Lat == nil || (*record.Lat != *msg.Lat)) {
		record.Lat = msg.Lat
		changed = true
	}

	if msg.Lng != nil && (record.Lng == nil || (*record.Lng != *msg.Lng)) {
		record.Lng = msg.Lng
		changed = true
	}

	if msg.Altitude != nil && (record.Altitude == nil || (*record.Altitude != *msg.Altitude)) {
		record.Altitude = msg.Altitude
		changed = true
	}

	if msg.GroundSpeed != nil && (record.GroundSpeed == nil || (*record.GroundSpeed != *msg.GroundSpeed)) {
		record.GroundSpeed = msg.GroundSpeed
		changed = true
	}

	if msg.Track != nil && (record.Track == nil || (*record.Track != *msg.Track)) {
		record.Track = msg.Track
		changed = true
	}

	if msg.VerticalRate != nil && (record.VerticalRate == nil || (*record.VerticalRate != *msg.VerticalRate)) {
		record.VerticalRate = msg.VerticalRate
		changed = true
	}

	if msg.Squawk != nil && (record.Squawk == nil || (*record.Squawk != *msg.Squawk)) {
		record.Squawk = msg.Squawk
		changed = true
	}

	now := int64(time.Now().UnixNano() / int64(time.Millisecond))
	record.Timestamp = &now

	if changed {
		updates <- record
	}

	if existingRecord != nil {
		proto.Merge(existingRecord, record)
		trackedCache.Add(ident, existingRecord)
	} else {
		trackedCache.Add(ident, record)
	}

	return nil
}

func parseMessage(message string) (*pb.SBS1Message, error) {
	parts := strings.Split(strings.TrimSuffix(message, "\n"), ",")

	if len(parts) != 22 {
		return nil, fmt.Errorf("Invalid message: %v", message)
	}

	var messageType pb.SBS1Message_MessageType
	switch parts[0] {
	case "SEL":
		messageType = pb.SBS1Message_SELECTION_CHANGE
		break
	case "ID":
		messageType = pb.SBS1Message_NEW_ID
		break
	case "AIR":
		messageType = pb.SBS1Message_NEW_AIRCRAFT
		break
	case "STA":
		messageType = pb.SBS1Message_STATUS_AIRCRAFT
		break
	case "CLK":
		messageType = pb.SBS1Message_CLICK
		break
	case "MSG":
		messageType = pb.SBS1Message_TRANSMISSION
		break
	}

	transmissionType, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("Invalid transmissionType: %v", parts[1])
	}

	altitude, err := safeParseInt(parts[11])
	if err != nil {
		return nil, fmt.Errorf("Invalid altitude: %v", parts[11])
	}

	groundSpeed, err := safeParseInt(parts[12])
	if err != nil {
		return nil, fmt.Errorf("Invalid groundSpeed: %v", parts[12])
	}

	track, err := safeParseInt(parts[13])
	if err != nil {
		return nil, fmt.Errorf("Invalid track: %v", parts[13])
	}

	lat, err := safeParseFloat(parts[14])
	if err != nil {
		return nil, fmt.Errorf("Invalid lat: %v", parts[14])
	}

	lng, err := safeParseFloat(parts[15])
	if err != nil {
		return nil, fmt.Errorf("Invalid lng: %v", parts[15])
	}

	verticalRate, err := safeParseInt(parts[16])
	if err != nil {
		return nil, fmt.Errorf("Invalid verticalRate: %v", parts[16])
	}

	transmissionTypePtr := pb.SBS1Message_TransmissionType(transmissionType)

	p := pb.SBS1Message{
		MessageType:      &messageType,
		TransmissionType: &transmissionTypePtr,
		SessionId:        nilableString(parts[2]),
		AircraftId:       nilableString(parts[3]),
		HexIdent:         nilableString(parts[4]),
		FlightId:         nilableString(parts[5]),
		GeneratedDate:    nilableString(parts[6]),
		GeneratedTime:    nilableString(parts[7]),
		LoggedDate:       nilableString(parts[8]),
		LoggedTime:       nilableString(parts[9]),
		Callsign:         nilableString(parts[10]),
		Altitude:         altitude,
		GroundSpeed:      groundSpeed,
		Track:            track,
		Lat:              lat,
		Lng:              lng,
		VerticalRate:     verticalRate,
		Squawk:           nilableString(parts[17]),
	}

	return &p, nil
}

func startSbsTracking(trackedCache *lru.Cache, updates chan<- *pb.PositionUpdate) {
	if *dumpHost == "" {
		logger.Debug("No dump host specified, no updates will be posted")
		return
	}

	b := &backoff.Backoff{
		Min:    500 * time.Millisecond,
		Max:    10 * time.Second,
		Factor: 2,
		Jitter: false,
	}

	for {
		logger.Infof("Connecting to %s", *dumpHost)

		c, err := net.Dial("tcp", *dumpHost)
		if err != nil {
			logger.Errorf("Error connecting to %s: %v", *dumpHost, err)
			sleepAmount := b.Duration()
			logger.Errorf("Retrying in %s", sleepAmount)
			time.Sleep(sleepAmount)
			continue
		}
		defer c.Close()

		reader := bufio.NewReader(c)
		b.Reset()

		for {
			message, err := reader.ReadString('\n')
			if err != nil {
				logger.Errorf("Error reading from host: %s", err)
				break
			}
			err = processMessage(message, trackedCache, updates)
			if err != nil {
				logger.Errorf("Error processing message: %s", err)
				break
			}
		}
	}
}

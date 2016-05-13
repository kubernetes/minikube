package goryman

import (
	"fmt"
	"os"
	"reflect"
	"time"

	pb "github.com/golang/protobuf/proto"
	"github.com/bigdatadev/goryman/proto"
)

// EventToProtocolBuffer converts an Event type to a proto.Event
func EventToProtocolBuffer(event *Event) (*proto.Event, error) {
	if event.Host == "" {
		event.Host, _ = os.Hostname()
	}
	if event.Time == 0 {
		event.Time = time.Now().Unix()
	}

	var e proto.Event
	t := reflect.ValueOf(&e).Elem()
	s := reflect.ValueOf(event).Elem()
	typeOfEvent := s.Type()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		value := reflect.ValueOf(f.Interface())
		if reflect.Zero(f.Type()) != value && f.Interface() != nil {
			name := typeOfEvent.Field(i).Name
			switch name {
			case "State", "Service", "Host", "Description":
				tmp := reflect.ValueOf(pb.String(value.String()))
				t.FieldByName(name).Set(tmp)
			case "Ttl":
				tmp := reflect.ValueOf(pb.Float32(float32(value.Float())))
				t.FieldByName(name).Set(tmp)
			case "Time":
				tmp := reflect.ValueOf(pb.Int64(value.Int()))
				t.FieldByName(name).Set(tmp)
			case "Tags":
				tmp := reflect.ValueOf(value.Interface().([]string))
				t.FieldByName(name).Set(tmp)
			case "Metric":
				switch reflect.TypeOf(f.Interface()).Kind() {
				case reflect.Int:
					tmp := reflect.ValueOf(pb.Int64(int64(value.Int())))
					t.FieldByName("MetricSint64").Set(tmp)
				case reflect.Float32:
					tmp := reflect.ValueOf(pb.Float32(float32(value.Float())))
					t.FieldByName("MetricF").Set(tmp)
				case reflect.Float64:
					tmp := reflect.ValueOf(pb.Float64(value.Float()))
					t.FieldByName("MetricD").Set(tmp)
				default:
					return nil, fmt.Errorf("Metric of invalid type (type %v)",
						reflect.TypeOf(f.Interface()).Kind())
				}
			case "Attributes":
				var attrs []*proto.Attribute
				for k, v := range value.Interface().(map[string]string) {
					k_, v_ := k, v
					attrs = append(attrs, &proto.Attribute{
						Key:   &k_,
						Value: &v_,
					})
				}
				t.FieldByName(name).Set(reflect.ValueOf(attrs))
			}
		}
	}
	return &e, nil
}

// StateToProtocolBuffer converts a State type to a proto.State
func StateToProtocolBuffer(state *State) (*proto.State, error) {
	if state.Host == "" {
		state.Host, _ = os.Hostname()
	}
	if state.Time == 0 {
		state.Time = time.Now().Unix()
	}

	var e proto.State
	t := reflect.ValueOf(&e).Elem()
	s := reflect.ValueOf(state).Elem()
	typeOfEvent := s.Type()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		value := reflect.ValueOf(f.Interface())
		if reflect.Zero(f.Type()) != value && f.Interface() != nil {
			name := typeOfEvent.Field(i).Name
			switch name {
			case "State", "Service", "Host", "Description":
				tmp := reflect.ValueOf(pb.String(value.String()))
				t.FieldByName(name).Set(tmp)
			case "Once":
				tmp := reflect.ValueOf(pb.Bool(bool(value.Bool())))
				t.FieldByName(name).Set(tmp)
			case "Ttl":
				tmp := reflect.ValueOf(pb.Float32(float32(value.Float())))
				t.FieldByName(name).Set(tmp)
			case "Time":
				tmp := reflect.ValueOf(pb.Int64(value.Int()))
				t.FieldByName(name).Set(tmp)
			case "Tags":
				tmp := reflect.ValueOf(value.Interface().([]string))
				t.FieldByName(name).Set(tmp)
			case "Metric":
				switch reflect.TypeOf(f.Interface()).Kind() {
				case reflect.Int:
					tmp := reflect.ValueOf(pb.Int64(int64(value.Int())))
					t.FieldByName("MetricSint64").Set(tmp)
				case reflect.Float32:
					tmp := reflect.ValueOf(pb.Float32(float32(value.Float())))
					t.FieldByName("MetricF").Set(tmp)
				case reflect.Float64:
					tmp := reflect.ValueOf(pb.Float64(value.Float()))
					t.FieldByName("MetricD").Set(tmp)
				default:
					return nil, fmt.Errorf("Metric of invalid type (type %v)",
						reflect.TypeOf(f.Interface()).Kind())
				}
			case "Attributes":
				var attrs []*proto.Attribute
				for k, v := range value.Interface().(map[string]string) {
					k_, v_ := k, v
					attrs = append(attrs, &proto.Attribute{
						Key:   &k_,
						Value: &v_,
					})
				}
				t.FieldByName(name).Set(reflect.ValueOf(attrs))
			}
		}
	}
	return &e, nil
}

// ProtocolBuffersToEvents converts an array of proto.Event to an array of Event
func ProtocolBuffersToEvents(pbEvents []*proto.Event) []Event {
	var events []Event
	for _, event := range pbEvents {
		e := Event{
			State:       event.GetState(),
			Service:     event.GetService(),
			Host:        event.GetHost(),
			Description: event.GetDescription(),
			Ttl:         event.GetTtl(),
			Time:        event.GetTime(),
			Tags:        event.GetTags(),
		}
		if event.MetricF != nil {
			e.Metric = event.GetMetricF()
		} else if event.MetricD != nil {
			e.Metric = event.GetMetricD()
		} else {
			e.Metric = event.GetMetricSint64()
		}
		if event.Attributes != nil {
			e.Attributes = make(map[string]string, len(event.GetAttributes()))
			for _, attr := range event.GetAttributes() {
				e.Attributes[attr.GetKey()] = attr.GetValue()
			}
		}
		events = append(events, e)
	}
	return events
}

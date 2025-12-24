# opening_hours.go

Go implementation of the OpenStreetMap opening_hours specification.

## Installation

go get github.com/Daquisu/opening_hours.go

## Usage

oh, err := openinghours.New("Mo-Fr 09:00-17:00")
if err != nil {
    // handle error
}
open := oh.GetState(time.Now())

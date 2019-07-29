package applehc

import (
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/service"
)

type Humidity struct {
	*accessory.Accessory

	HumiditySensor *service.HumiditySensor
}

func NewHumiditySensor(info accessory.Info, humidity, min, max, steps float64) *Humidity {
	acc := Humidity{}
	acc.Accessory = accessory.New(info, accessory.TypeSensor)
	acc.HumiditySensor = service.NewHumiditySensor()
	acc.HumiditySensor.CurrentRelativeHumidity.SetValue(humidity)
	acc.HumiditySensor.CurrentRelativeHumidity.SetMinValue(min)
	acc.HumiditySensor.CurrentRelativeHumidity.SetMaxValue(max)
	acc.HumiditySensor.CurrentRelativeHumidity.SetStepValue(steps)

	acc.AddService(acc.HumiditySensor.Service)

	return &acc
}

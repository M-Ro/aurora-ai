package stablediffusion

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"

	"github.com/M-Ro/aurora-ai/api"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type ParameterSet struct {
	TaskId            string
	PositivePrompt    string
	NegativePrompt    string
	DunnoA            []string
	SampleSteps       uint32
	SampleMethod      string
	RestoreFaces      bool
	Tiling            bool
	BatchCount        uint32
	BatchSize         uint32
	CfgScale          float64
	Seed              int32
	DunnoB            int32
	DunnoC            int32
	DunnoD            int32
	DunnoE            int32
	DunnoF            bool
	Height            uint32
	Width             uint32
	HiresFix          bool
	DenoisingStrength float64
	UpscaleBy         float64
	Upscaler          string
	HiresSteps        uint32
	ResizeWidthTo     uint32
	ResizeHeightTo    uint32
	DunnoG            []string
	Script            string
	DunnoH            bool
	DunnoI            bool
	DunnoJ            string   // "positive"
	DunnoK            string   // "comma"
	DunnoL            uint     // 0
	DunnoM            bool     // false
	DunnoN            bool     // false
	DunnoO            string   // ""
	XType             string   // "Seed"
	DunnoP            string   // ""
	DunnoQ            string   // "Nothing"
	DunnoR            string   // ""
	DunnoS            string   // "Nothing"
	DunnoT            string   // ""
	DunnoU            bool     // true
	DunnoW            bool     // false
	DunnoX            bool     // false
	DunnoY            bool     // false
	DunnoZ            uint32   // 0
	DunnoAA           []string // []
	DunnoAB           string   // ""
	DunnoAC           string   // ""
	DunnoAD           string   // ""
}

func (p ParameterSet) MarshalJSON() ([]byte, error) {
	v := reflect.ValueOf(p)
	refValues := make([]interface{}, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		refValues[i] = v.Field(i).Interface()
	}

	return json.Marshal(refValues)
}

func generateTaskId() string {
	return fmt.Sprintf("task(%d)", rand.Int63())
}

func NewParameterSet() ParameterSet {
	return ParameterSet{
		TaskId:            generateTaskId(),
		PositivePrompt:    "",
		NegativePrompt:    "",
		DunnoA:            []string{},
		SampleSteps:       uint32(viper.GetInt("stable_diffusion.defaults.sample_steps")),
		SampleMethod:      viper.GetString("stable_diffusion.defaults.sampler"),
		RestoreFaces:      false,
		Tiling:            false,
		BatchCount:        1,
		BatchSize:         uint32(viper.GetInt("stable_diffusion.defaults.batch_size")),
		CfgScale:          viper.GetFloat64("stable_diffusion.defaults.cfg_scale"),
		Seed:              viper.GetInt32("stable_diffusion.defaults.seed"),
		DunnoB:            -1,
		DunnoC:            0,
		DunnoD:            0,
		DunnoE:            0,
		DunnoF:            false,
		Height:            uint32(viper.GetInt("stable_diffusion.defaults.height")),
		Width:             uint32(viper.GetInt("stable_diffusion.defaults.width")),
		HiresFix:          false,
		DenoisingStrength: 0.7,
		UpscaleBy:         2,
		Upscaler:          "Latent",
		HiresSteps:        0,
		ResizeWidthTo:     0,
		ResizeHeightTo:    0,
		DunnoG:            []string{},
		Script:            "None",
		DunnoH:            false,
		DunnoI:            false,
		DunnoJ:            "positive",
		DunnoK:            "comma",
		DunnoL:            0,
		DunnoM:            false,
		DunnoN:            false,
		DunnoO:            "",
		XType:             "Seed",
		DunnoP:            "",
		DunnoQ:            "Nothing",
		DunnoR:            "",
		DunnoS:            "Nothing",
		DunnoT:            "",
		DunnoU:            true,
		DunnoW:            false,
		DunnoX:            false,
		DunnoY:            false,
		DunnoZ:            0,
		DunnoAA:           []string{},
		DunnoAB:           "",
		DunnoAC:           "",
		DunnoAD:           "",
	}
}

type SdResponsePacket struct {
	Message api.GradioResponseMessage `json:"msg"`
	Output  *SdResponseOutput         `json:"output"`
	Success *bool                     `json:"success"`
}

type SdResponseOutput struct {
	Data            DataBlock `json:"data"`
	IsGenerating    bool      `json:"is_generating"`
	Duration        float32   `json:"duration"`
	AverageDuration float32   `json:"average_duration"`
}

type DataBlock struct {
	Images []ImageBlock
	// we don't care about the rest
}

type ImageBlock struct {
	Filename string  `json:"name"`
	IsFile   bool    `json:"is_file"`
	Data     *string `json:"data"`
}

func (i *DataBlock) UnmarshalJSON(data []byte) error {
	var v []interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		logrus.Error("Couldn't even unmarshal into a fecking thingy")
		return err
	}

	imageBlocks := []ImageBlock{}

	imageMaps, _ := v[0].([]interface{})
	for _, imgMap := range imageMaps {
		v := imgMap.(map[string]interface{})

		imageBlocks = append(imageBlocks, ImageBlock{
			Filename: v["name"].(string),
			IsFile:   v["is_file"].(bool),
			Data:     nil, // Does not appear to be used by the API
		})
	}

	i.Images = imageBlocks

	return nil
}

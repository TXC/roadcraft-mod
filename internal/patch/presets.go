package patch

// GetPresetPatches returns predefined patches for common vehicle modifications
func GetPresetPatches() []VehiclePatch {
	return []VehiclePatch{
		{
			VehicleName: "Zikz 605E Mobile Scalper",
			FilePattern: "ssl/autogen_designer_wizard/trucks/auto_zikz_605e_mobile_scalper_res/auto_zikz_605e_mobile_scalper_res.cls",
			Patches: []Patch{
				{
					Name:      "Remove sand percentage requirement",
					Path:      "properties.prop_truck_mobile_sand_screen.allowedPercent",
					Value:     "0.0", // Was 0.4
					Operation: OpSet,
				},
				{
					Name:      "Increase sand distance",
					Path:      "properties.prop_truck_mobile_sand_screen.sandDistance",
					Value:     "600", // Was 150
					Operation: OpSet,
				},
				{
					Name:      "Increase sand distance",
					Path:      "properties.prop_usable.smartsEntryPoints.SandStorage.checkers.UsableCheckerDistance.distance",
					Value:     "600", // Was 150
					Operation: OpSet,
				},
				{
					Name:      "Increase sand focus distance",
					Path:      "properties.prop_usable.smartsEntryPoints.SandStorage.focusDistance",
					Value:     "600", // Was 150
					Operation: OpSet,
				},
			},
		},
		{
			VehicleName: "Aramatsu Bowhead 60T",
			FilePattern: "ssl/autogen_designer_wizard/trucks/auto_aramatsu_bowhead_heavy_dumptruck_new/auto_aramatsu_bowhead_heavy_dumptruck_new.cls",
			Patches: []Patch{
				{
					Name:      "Increase front suspensionSet strength",
					Path:      "properties.prop_truck_rb.suspensionSet.front.strength",
					Value:     "2.4", // Was 0.8
					Operation: OpSet,
				},
				{
					Name:      "Increase rear suspensionSet strength",
					Path:      "properties.prop_truck_rb.suspensionSet.rear.strength",
					Value:     "0.32", // Was 0.08
					Operation: OpSet,
				},
				{
					Name:      "Increase suspensionAxleCollection[0] strength",
					Path:      "properties.prop_truck_rb.suspensionAxleCollection[0].strength",
					Value:     "2.4", // Was 0.8
					Operation: OpSet,
				},
				{
					Name:      "Increase suspensionAxleCollection[1] strength",
					Path:      "properties.prop_truck_rb.suspensionAxleCollection[1].strength",
					Value:     "0.32", // Was 0.08
					Operation: OpSet,
				},
				{
					Name:      "Increase suspensionAxleCollection[2] strength",
					Path:      "properties.prop_truck_rb.suspensionAxleCollection[2].strength",
					Value:     "0.32", // Was 0.08
					Operation: OpSet,
				},
				{
					Name:      "Increase engine torque",
					Path:      "properties.prop_truck_rb.engine.params.torque",
					Value:     "900000", // Was 600000
					Operation: OpSet,
				},
				{
					Name: "Add 4th gear to transmission",
					Path: "properties.prop_truck_rb.gearbox.params.gears",
					Value: map[string]any{
						"angularVelocity": "12",
						"gearRatio":       "1.5",
						"targetSpeedRangeKmh": map[string]any{
							"x":     "45",
							"y":     "70",
							"_type": "Vector2dRnd",
						},
					},
					Operation: OpAdd,
				},
				{
					Name: "Add 5th gear to transmission",
					Path: "properties.prop_truck_rb.gearbox.params.gears",
					Value: map[string]any{
						"angularVelocity": "16",
						"gearRatio":       "1.0",
						"targetSpeedRangeKmh": map[string]any{
							"x":     "70",
							"y":     "95",
							"_type": "Vector2dRnd",
						},
					},
					Operation: OpAdd,
				},
				{
					Name:      "Change WheelSubstanceResistanceMult[1]",
					Path:      "properties.prop_truck_rb.wheelsSubstanceResistanceMult[1]",
					Value:     "0.2", // Was 0.4
					Operation: OpSet,
				},
				{
					Name:      "Change WheelSubstanceResistanceMult[3]",
					Path:      "properties.prop_truck_rb.wheelsSubstanceResistanceMult[3]",
					Value:     "0.25", // Was 0.45
					Operation: OpSet,
				},
				{
					Name:      "Change WheelSubstanceResistanceMult[5]",
					Path:      "properties.prop_truck_rb.wheelsSubstanceResistanceMult[5]",
					Value:     "0.25", // Was 0.45
					Operation: OpSet,
				},
				{
					Name:      "Change WaterDragMultiplier",
					Path:      "properties.prop_phys_substance_target.waterDragMultiplier",
					Value:     "1.25", // Was 4.25
					Operation: OpSet,
				},
				{
					Name:      "Change Load Mass of the vehicle",
					Path:      "properties.prop_load_volume.volumeMass",
					Value:     "80000", // Was 20000
					Operation: OpSet,
				},
			},
		},
		{
			VehicleName: "Baikal 65206 60T",
			FilePattern: "ssl/autogen_designer_wizard/trucks/auto_baikal_65206_heavy_dumptruck_res/auto_baikal_65206_heavy_dumptruck_res.cls",
			Patches: []Patch{
				{
					Name:      "Change Load Mass of the vehicle",
					Path:      "properties.prop_load_volume.volumeMass",
					Value:     "80000", // Was 40000
					Operation: OpSet,
				},
			},
		},
	}
}

// GetPatchByName returns a specific preset patch by name
func GetPatchByName(name string) *VehiclePatch {
	for _, p := range GetPresetPatches() {
		if p.VehicleName == name {
			return &p
		}
	}
	return nil
}

// ListPresets returns a list of all available preset names
func ListPresets() []string {
	presets := GetPresetPatches()
	names := make([]string, len(presets))
	for i, p := range presets {
		names[i] = p.VehicleName
	}
	return names
}

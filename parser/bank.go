package parser

import (
	"bytes"
	wio "dekr0/hd2_audio_db/io"
	"io"
)

type HircType uint8

const (
	HircTypeState           HircType = 0x01
	HircTypeSound           HircType = 0x02
	HircTypeAction          HircType = 0x03
	HircTypeEvent           HircType = 0x04
	HircTypeRanSeqCntr      HircType = 0x05
	HircTypeSwitchCntr      HircType = 0x06
	HircTypeActorMixer      HircType = 0x07
	HircTypeBus             HircType = 0x08
	HircTypeLayerCntr       HircType = 0x09
	HircTypeMusicSegment    HircType = 0x0a
	HircTypeMusicTrack      HircType = 0x0b
	HircTypeMusicSwitchCntr HircType = 0x0c
	HircTypeMusicRanSeqCntr HircType = 0x0d
	HircTypeAttenuation     HircType = 0x0e
	HircTypeDialogueEvent   HircType = 0x0f
	HircTypeFxShareSet      HircType = 0x10
	HircTypeFxCustom        HircType = 0x11
	HircTypeAuxBus          HircType = 0x12
	HircTypeLFOModulator    HircType = 0x13
	HircEnvelopeModulator   HircType = 0x14
	HircAudioDevice         HircType = 0x15
	HircTimeModulator       HircType = 0x16
)

var HircTypeName []string = []string{
	"",
	"State",
	"Sound",
	"Action",
	"Event",
	"Random / Sequence Container",
	"Switch Container",
	"Actor Mixer",
	"Bus",
	"Layer Container",
	"Music Segment",
	"Music Track",
	"Music Switch Container",
	"Music Random / Sequence Container",
	"Attenuation",
	"Dialogue Event",
    "FX Share Set",
    "FX Custom",
    "Aux Bus",
    "LFO Modulator",
    "Envelope Modulator",
    "Audio Device",
    "Time Modulator",
}

func ParseBank(r *wio.Reader, end uint64) *HIRC {
	tag := make([]byte, 4, 4)
	var size uint32
	var err error
	for r.Tell() <= uint(end) {
		err = r.ReadFull(tag)
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		size = r.U32Unsafe()
		if bytes.Compare(tag, []byte{'H', 'I', 'R', 'C'}) != 0 {
			r.RelSeekUnsafe(int(size))
			continue
		}
		end := r.Tell() + uint(size)
		hirc := parseHIRC(r)
		if r.Tell() != end {
			panic("Reader position does not end up at expected location after parsing HIRC.")
		}
		return hirc
	}
	return nil
}

func parseHIRC(r *wio.Reader) *HIRC {
	n := r.U32Unsafe()

	hirc := HIRC{
		Header: n,
		Hierarchy: make([]Hierarchy, n, n),
		Sound: make([]Sound, 0, n / 2),
	}

	for i := range n {
		t := r.U8Unsafe()
		size := r.U32Unsafe()

		hirc.Hierarchy[i].Type = HircType(t)
		switch hirc.Hierarchy[i].Type {
		case HircTypeState:
			parseState(r, size, i, hirc.Hierarchy)
		case HircTypeSound:
			hirc.Sound = parseSound(r, size, i, hirc.Hierarchy, hirc.Sound)
		case HircTypeAction:
			parseAction(r, size, i, hirc.Hierarchy)
		case HircTypeEvent:
			parseEvent(r, size, i, hirc.Hierarchy)
		case HircTypeRanSeqCntr:
			parseRanSeqCntr(r, size, i, hirc.Hierarchy)
		case HircTypeSwitchCntr:
			parseSwitchCntr(r, size, i, hirc.Hierarchy)
		case HircTypeActorMixer:
			parseActorMixer(r, size, i, hirc.Hierarchy)
		case HircTypeBus:
			parseBus(r, size, i, hirc.Hierarchy)
		case HircTypeLayerCntr:
			parseLayerCntr(r, size, i, hirc.Hierarchy)
		case HircTypeMusicSegment:
			parseMusicSegment(r, size, i, hirc.Hierarchy)
		case HircTypeMusicTrack:
			hirc.Sound = parseMusicTrack(r, size, i, hirc.Hierarchy, hirc.Sound)
		case HircTypeMusicSwitchCntr:
			parseMusicSwitchCntr(r, size, i, hirc.Hierarchy)
		case HircTypeMusicRanSeqCntr:
			parseMusicRanSeqCntr(r, size, i, hirc.Hierarchy)
		case HircTypeAttenuation:
			parseAttenuation(r, size, i, hirc.Hierarchy)
		case HircTypeDialogueEvent:
			parseDialogueEvent(r, size, i, hirc.Hierarchy)
		case HircTypeFxShareSet:
			parseFxShareSet(r, size, i, hirc.Hierarchy)
		case HircTypeFxCustom:
			parseFxShareCustom(r, size, i, hirc.Hierarchy)
		case HircTypeAuxBus:
			parseAuxBus(r, size, i, hirc.Hierarchy)
		case HircTypeLFOModulator:
			parseLFOModulator(r, size, i, hirc.Hierarchy)
		case HircEnvelopeModulator:
			parseEnvelopeModulator(r, size, i, hirc.Hierarchy)
		case HircAudioDevice:
			parseAudioDevice(r, size, i, hirc.Hierarchy)
		case HircTimeModulator:
			parseTimeModulator(r, size, i, hirc.Hierarchy)
		}
	}
	return &hirc
}

func parseState(r *wio.Reader, size uint32, i uint32, hirc []Hierarchy) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.AbsSeekUnsafe(end)
}

func parseSound(
	r *wio.Reader,
	size uint32,
	i uint32,
	hirc []Hierarchy,
	sounds []Sound,
) []Sound {
	begin := r.Tell()
	end := begin + uint(size)

	hirc[i].ID = r.U32Unsafe()

	sound := Sound{Idx: i}

	parseBankSourceData(r, &sound)

	sounds = append(sounds, sound)

	hirc[i].Parent = parseBaseParam(r)

	r.AbsSeekUnsafe(end)

	return sounds
}

func parseBankSourceData(r *wio.Reader, sound *Sound) {
	sound.PluginID = r.U32Unsafe()
	sound.StreamType = r.U8Unsafe()
	sound.SourceID = r.U32Unsafe()
	sound.InMemoryMediaSize = r.U32Unsafe()
	sound.SourceBits = r.U8Unsafe()
	sound.PluginParamSize = 0

	hasParam := (sound.PluginID & 0x0F) == 2 && 
	            (sound.PluginID != 0)
	if hasParam {
		sound.PluginParamSize = r.U32Unsafe()
		if sound.PluginParamSize > 0 {
			r.RelSeekUnsafe(int(sound.PluginParamSize))
		}
	}
}

func parseBaseParam(r *wio.Reader) uint32 {
	r.RelSeekUnsafe(1) // BitIsOverrideParentFx

	// FxChunk
	uniqueNumFX := r.U8Unsafe()
	if uniqueNumFX > 0 {
		r.RelSeekUnsafe(1)
		r.RelSeekUnsafe(int(uniqueNumFX) * 7)
	}

	// FxChunkMetadata
	r.RelSeekUnsafe(1)
	uniqueNumFXMetadata := r.U8Unsafe()
	r.RelSeekUnsafe(int(uniqueNumFXMetadata) * 6)

	r.RelSeekUnsafe(1 + 4) // BitOverrideAttachmentParams + OverrideBusId

	return r.U32Unsafe()
}

func parseAction(r *wio.Reader, size uint32, i uint32, hirc []Hierarchy) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.RelSeekUnsafe(2) // ulActionType
	hirc[i].Parent = r.U32Unsafe() // idExt
	r.AbsSeekUnsafe(end)
}

func parseEvent(r *wio.Reader, size uint32, i uint32, hirc []Hierarchy) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.AbsSeekUnsafe(end)
}

func parseRanSeqCntr(r *wio.Reader, size uint32, i uint32, hirc []Hierarchy) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	hirc[i].Parent = parseBaseParam(r)
	r.AbsSeekUnsafe(end)
}

func parseSwitchCntr(r *wio.Reader, size uint32, i uint32, hirc []Hierarchy) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	hirc[i].Parent = parseBaseParam(r)
	r.AbsSeekUnsafe(end)
}

func parseActorMixer(r *wio.Reader, size uint32, i uint32, hirc []Hierarchy) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	hirc[i].Parent = parseBaseParam(r)
	r.AbsSeekUnsafe(end)
}

func parseBus(r *wio.Reader, size uint32, i uint32, hirc []Hierarchy) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.AbsSeekUnsafe(end)
}

func parseLayerCntr(r *wio.Reader, size uint32, i uint32, hirc []Hierarchy) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	hirc[i].Parent = parseBaseParam(r)
	r.AbsSeekUnsafe(end)
}

func parseMusicSegment(r *wio.Reader, size uint32, i uint32, hirc []Hierarchy) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.RelSeekUnsafe(1) // uFlags
	hirc[i].Parent = parseBaseParam(r)
	r.AbsSeekUnsafe(end)
}

func parseMusicTrack(
	r *wio.Reader,
	size uint32,
	i uint32,
	hirc []Hierarchy,
	sounds []Sound,
) []Sound {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()

	r.RelSeekUnsafe(1) // uFlags

	numSources := r.U32Unsafe()
	for range numSources {
		sound := Sound{Idx: i}
		parseBankSourceData(r, &sound)
		sounds = append(sounds, sound)
	}

	numPlayListItem := r.U32Unsafe()
	// trackID
	// sourceID
	// eventID
	// fPlayAt
	// fBeginTrimOffset
	// fEndTrimOffset
	// fSrcDuration
	r.RelSeekUnsafe(int(numPlayListItem) * (3 * 4 + 4 * 8))
	if numPlayListItem > 0 {
		r.RelSeekUnsafe(4) // numSubTrack
	}

	numClipAutomationItem := r.U32Unsafe()
	for range numClipAutomationItem {
		// uClipIndex
		// eAutoType
		r.RelSeekUnsafe(2 * 4)
		// uNumPoints
		// From
		// To
		// Interp
		r.RelSeekUnsafe(int(r.U32Unsafe()) * (3 * 4))
	}

	hirc[i].Parent = parseBaseParam(r)

	r.AbsSeekUnsafe(end)

	return sounds
}

func parseMusicSwitchCntr(
	r *wio.Reader,
	size uint32,
	i uint32,
	hirc []Hierarchy,
) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.RelSeekUnsafe(1) // uFlags
	hirc[i].Parent = parseBaseParam(r)
	r.AbsSeekUnsafe(end)
}

func parseMusicRanSeqCntr(
	r *wio.Reader,
	size uint32,
	i uint32,
	hirc []Hierarchy,
) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.RelSeekUnsafe(1) // uFlags
	hirc[i].Parent = parseBaseParam(r)
	r.AbsSeekUnsafe(end)
}

func parseAttenuation(
	r *wio.Reader,
	size uint32,
	i uint32,
	hirc []Hierarchy,
) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.AbsSeekUnsafe(end)
}

func parseDialogueEvent(
	r *wio.Reader,
	size uint32,
	i uint32,
	hirc []Hierarchy,
) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.AbsSeekUnsafe(end)
}

func parseFxShareSet(
	r *wio.Reader,
	size uint32,
	i uint32,
	hirc []Hierarchy,
) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.AbsSeekUnsafe(end)
}

func parseFxShareCustom(
	r *wio.Reader,
	size uint32,
	i uint32,
	hirc []Hierarchy,
) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.AbsSeekUnsafe(end)
}

func parseAuxBus(
	r *wio.Reader,
	size uint32,
	i uint32,
	hirc []Hierarchy,
) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.AbsSeekUnsafe(end)
}

func parseLFOModulator(
	r *wio.Reader,
	size uint32,
	i uint32,
	hirc []Hierarchy,
) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.AbsSeekUnsafe(end)
}

func parseEnvelopeModulator(
	r *wio.Reader,
	size uint32,
	i uint32,
	hirc []Hierarchy,
) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.AbsSeekUnsafe(end)
}

func parseAudioDevice(
	r *wio.Reader,
	size uint32,
	i uint32,
	hirc []Hierarchy,
) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.AbsSeekUnsafe(end)
}

func parseTimeModulator(
	r *wio.Reader,
	size uint32,
	i uint32,
	hirc []Hierarchy,
) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.AbsSeekUnsafe(end)
}

func ParseBankInPlace(r *wio.InPlaceReader, end uint64) *HIRC {
	for r.Len() > 0 {
		tag, err := r.FourCC()
		if err != nil {
			break
		}
		size := r.U32Unsafe()
		if bytes.Compare(tag, []byte{'H', 'I', 'R', 'C'}) != 0 {
			if err := r.RelSeek(int(size)); err != nil {
				break
			}
			continue
		}
		hirc := parseHIRCInPlace(r)
		return hirc
	}
	return nil
}

func parseHIRCInPlace(r *wio.InPlaceReader) *HIRC {
	n := r.U32Unsafe()

	hirc := HIRC{
		Hierarchy: make([]Hierarchy, n, n),
		Sound: make([]Sound, 0, n / 2),
	}

	for i := range n {
		t := r.U8Unsafe()
		size := r.U32Unsafe()

		hirc.Hierarchy[i].Type = HircType(t)
		switch hirc.Hierarchy[i].Type {
		case HircTypeState:
			parseStateInPlace(r, size, i, hirc.Hierarchy)
		case HircTypeSound:
			hirc.Sound = parseSoundInPlace(r, size, i, hirc.Hierarchy, hirc.Sound)
		case HircTypeAction:
			parseActionInPlace(r, size, i, hirc.Hierarchy)
		case HircTypeEvent:
			parseEventInPlace(r, size, i, hirc.Hierarchy)
		case HircTypeRanSeqCntr:
			parseRanSeqCntrInPlace(r, size, i, hirc.Hierarchy)
		case HircTypeSwitchCntr:
			parseSwitchCntrInPlace(r, size, i, hirc.Hierarchy)
		case HircTypeActorMixer:
			parseActorMixerInPlace(r, size, i, hirc.Hierarchy)
		case HircTypeBus:
			parseBusInPlace(r, size, i, hirc.Hierarchy)
		case HircTypeLayerCntr:
			parseLayerCntrInPlace(r, size, i, hirc.Hierarchy)
		case HircTypeMusicSegment:
			parseMusicSegmentInPlace(r, size, i, hirc.Hierarchy)
		case HircTypeMusicTrack:
			hirc.Sound = parseMusicTrackInPlace(r, size, i, hirc.Hierarchy, hirc.Sound)
		case HircTypeMusicSwitchCntr:
			parseMusicSwitchCntrInPlace(r, size, i, hirc.Hierarchy)
		case HircTypeMusicRanSeqCntr:
			parseMusicRanSeqCntrInPlace(r, size, i, hirc.Hierarchy)
		case HircTypeAttenuation:
			parseAttenuationInPlace(r, size, i, hirc.Hierarchy)
		case HircTypeDialogueEvent:
			parseDialogueEventInPlace(r, size, i, hirc.Hierarchy)
		case HircTypeFxShareSet:
			parseFxShareSetInPlace(r, size, i, hirc.Hierarchy)
		case HircTypeFxCustom:
			parseFxShareCustomInPlace(r, size, i, hirc.Hierarchy)
		case HircTypeAuxBus:
			parseAuxBusInPlace(r, size, i, hirc.Hierarchy)
		case HircTypeLFOModulator:
			parseLFOModulatorInPlace(r, size, i, hirc.Hierarchy)
		case HircEnvelopeModulator:
			parseEnvelopeModulatorInPlace(r, size, i, hirc.Hierarchy)
		case HircAudioDevice:
			parseAudioDeviceInPlace(r, size, i, hirc.Hierarchy)
		case HircTimeModulator:
			parseTimeModulatorInPlace(r, size, i, hirc.Hierarchy)
		}
	}
	return &hirc
}

func parseStateInPlace(r *wio.InPlaceReader, size uint32, i uint32, hirc []Hierarchy) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.AbsSeekUnsafe(end)
}

func parseSoundInPlace(
	r *wio.InPlaceReader,
	size uint32,
	i uint32,
	hirc []Hierarchy,
	sounds []Sound,
) []Sound {
	begin := r.Tell()
	end := begin + uint(size)

	hirc[i].ID = r.U32Unsafe()

	sound := Sound{Idx: i}

	parseBankSourceDataInPlace(r, &sound)

	sounds = append(sounds, sound)

	hirc[i].Parent = parseBaseParamInPlace(r)

	r.AbsSeekUnsafe(end)

	return sounds
}

func parseBankSourceDataInPlace(r *wio.InPlaceReader, sound *Sound) {
	sound.PluginID = r.U32Unsafe()
	sound.StreamType = r.U8Unsafe()
	sound.SourceID = r.U32Unsafe()
	sound.InMemoryMediaSize = r.U32Unsafe()
	sound.SourceBits = r.U8Unsafe()
	sound.PluginParamSize = 0

	hasParam := (sound.PluginID & 0x0F) == 2 && 
	            (sound.PluginID != 0)
	if hasParam {
		sound.PluginParamSize = r.U32Unsafe()
		if sound.PluginParamSize > 0 {
			r.RelSeekUnsafe(int(sound.PluginParamSize))
		}
	}
}

func parseBaseParamInPlace(r *wio.InPlaceReader) uint32 {
	r.RelSeekUnsafe(1) // BitIsOverrideParentFx

	// FxChunk
	uniqueNumFX := r.U8Unsafe()
	if uniqueNumFX > 0 {
		r.RelSeekUnsafe(1)
		r.RelSeekUnsafe(int(uniqueNumFX) * 7)
	}

	// FxChunkMetadata
	r.RelSeekUnsafe(1)
	uniqueNumFXMetadata := r.U8Unsafe()
	r.RelSeekUnsafe(int(uniqueNumFXMetadata) * 6)

	r.RelSeekUnsafe(1 + 4) // BitOverrideAttachmentParams + OverrideBusId

	return r.U32Unsafe()
}

func parseActionInPlace(r *wio.InPlaceReader, size uint32, i uint32, hirc []Hierarchy) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.RelSeekUnsafe(2) // ulActionType
	hirc[i].Parent = r.U32Unsafe() // idExt
	r.AbsSeekUnsafe(end)
}

func parseEventInPlace(r *wio.InPlaceReader, size uint32, i uint32, hirc []Hierarchy) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.AbsSeekUnsafe(end)
}

func parseRanSeqCntrInPlace(r *wio.InPlaceReader, size uint32, i uint32, hirc []Hierarchy) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	hirc[i].Parent = parseBaseParamInPlace(r)
	r.AbsSeekUnsafe(end)
}

func parseSwitchCntrInPlace(r *wio.InPlaceReader, size uint32, i uint32, hirc []Hierarchy) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	hirc[i].Parent = parseBaseParamInPlace(r)
	r.AbsSeekUnsafe(end)
}

func parseActorMixerInPlace(r *wio.InPlaceReader, size uint32, i uint32, hirc []Hierarchy) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	hirc[i].Parent = parseBaseParamInPlace(r)
	r.AbsSeekUnsafe(end)
}

func parseBusInPlace(r *wio.InPlaceReader, size uint32, i uint32, hirc []Hierarchy) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.AbsSeekUnsafe(end)
}

func parseLayerCntrInPlace(r *wio.InPlaceReader, size uint32, i uint32, hirc []Hierarchy) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	hirc[i].Parent = parseBaseParamInPlace(r)
	r.AbsSeekUnsafe(end)
}

func parseMusicSegmentInPlace(r *wio.InPlaceReader, size uint32, i uint32, hirc []Hierarchy) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.RelSeekUnsafe(1) // uFlags
	hirc[i].Parent = parseBaseParamInPlace(r)
	r.AbsSeekUnsafe(end)
}

func parseMusicTrackInPlace(
	r *wio.InPlaceReader,
	size uint32,
	i uint32,
	hirc []Hierarchy,
	sounds []Sound,
) []Sound {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()

	r.RelSeekUnsafe(1) // uFlags

	numSources := r.U32Unsafe()
	for range numSources {
		sound := Sound{Idx: i}
		parseBankSourceDataInPlace(r, &sound)
		sounds = append(sounds, sound)
	}

	numPlayListItem := r.U32Unsafe()
	// trackID
	// sourceID
	// eventID
	// fPlayAt
	// fBeginTrimOffset
	// fEndTrimOffset
	// fSrcDuration
	r.RelSeekUnsafe(int(numPlayListItem) * (3 * 4 + 4 * 8))
	if numPlayListItem > 0 {
		r.RelSeekUnsafe(4) // numSubTrack
	}

	numClipAutomationItem := r.U32Unsafe()
	for range numClipAutomationItem {
		// uClipIndex
		// eAutoType
		r.RelSeekUnsafe(2 * 4)
		// uNumPoints
		// From
		// To
		// Interp
		r.RelSeekUnsafe(int(r.U32Unsafe()) * (3 * 4))
	}

	hirc[i].Parent = parseBaseParamInPlace(r)

	r.AbsSeekUnsafe(end)

	return sounds
}

func parseMusicSwitchCntrInPlace(
	r *wio.InPlaceReader,
	size uint32,
	i uint32,
	hirc []Hierarchy,
) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.RelSeekUnsafe(1) // uFlags
	hirc[i].Parent = parseBaseParamInPlace(r)
	r.AbsSeekUnsafe(end)
}

func parseMusicRanSeqCntrInPlace(
	r *wio.InPlaceReader,
	size uint32,
	i uint32,
	hirc []Hierarchy,
) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.RelSeekUnsafe(1) // uFlags
	hirc[i].Parent = parseBaseParamInPlace(r)
	r.AbsSeekUnsafe(end)
}

func parseAttenuationInPlace(
	r *wio.InPlaceReader,
	size uint32,
	i uint32,
	hirc []Hierarchy,
) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.AbsSeekUnsafe(end)
}

func parseDialogueEventInPlace(
	r *wio.InPlaceReader,
	size uint32,
	i uint32,
	hirc []Hierarchy,
) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.AbsSeekUnsafe(end)
}

func parseFxShareSetInPlace(
	r *wio.InPlaceReader,
	size uint32,
	i uint32,
	hirc []Hierarchy,
) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.AbsSeekUnsafe(end)
}

func parseFxShareCustomInPlace(
	r *wio.InPlaceReader,
	size uint32,
	i uint32,
	hirc []Hierarchy,
) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.AbsSeekUnsafe(end)
}

func parseAuxBusInPlace(
	r *wio.InPlaceReader,
	size uint32,
	i uint32,
	hirc []Hierarchy,
) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.AbsSeekUnsafe(end)
}

func parseLFOModulatorInPlace(
	r *wio.InPlaceReader,
	size uint32,
	i uint32,
	hirc []Hierarchy,
) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.AbsSeekUnsafe(end)
}

func parseEnvelopeModulatorInPlace(
	r *wio.InPlaceReader,
	size uint32,
	i uint32,
	hirc []Hierarchy,
) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.AbsSeekUnsafe(end)
}

func parseAudioDeviceInPlace(
	r *wio.InPlaceReader,
	size uint32,
	i uint32,
	hirc []Hierarchy,
) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.AbsSeekUnsafe(end)
}

func parseTimeModulatorInPlace(
	r *wio.InPlaceReader,
	size uint32,
	i uint32,
	hirc []Hierarchy,
) {
	begin := r.Tell()
	end := begin + uint(size)
	hirc[i].ID = r.U32Unsafe()
	r.AbsSeekUnsafe(end)
}

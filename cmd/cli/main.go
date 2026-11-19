package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"personal-os/internal/audio"
	"personal-os/internal/chunker"
	"personal-os/internal/domain"
	"personal-os/internal/prompt"
	"personal-os/internal/synthesizer"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "generate":
		handleGenerate(os.Args[2:])
	case "version":
		fmt.Println("Personal OS Engine v0.1.0-alpha")
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Personal OS: AI Audio Motivation Engine")
	fmt.Println("\nUsage:")
	fmt.Println("  personal-os generate [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  --topic       Target theme for motivation monologue (required)")
	fmt.Println("  --persona     Vocal persona preset (default: elliot)")
	fmt.Println("  --tts-url     Remote/Cloud TTS endpoint URL (e.g. Hugging Face Space)")
	fmt.Println("  --ambient     Path to ambient background audio track (optional)")
	fmt.Println("  --output      Path for the final master MP3 audio output")
}

func handleGenerate(args []string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)

	topic := fs.String("topic", "conquering daily overthinking", "Motivation monologue theme")
	personaID := fs.String("persona", "elliot", "Vocal persona identifier")
	ttsURL := fs.String("tts-url", "", "Cloud or local TTS endpoint URL")
	ambientPath := fs.String("ambient", "", "Path to ambient background track (optional)")

	defaultOutput := filepath.Join("storage", "output", fmt.Sprintf("daily_%s.mp3", time.Now().Format("2006-01-02")))
	outputPath := fs.String("output", defaultOutput, "Destination path for master MP3")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse command flags: %v\n", err)
		os.Exit(1)
	}

	// 1. Resolve Persona configuration
	persona, err := domain.GetPersonaByID(*personaID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Persona error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("==================================================")
	fmt.Println("🎙️  PERSONAL OS - DAILY MOTIVATION ENGINE")
	fmt.Println("==================================================")
	fmt.Printf("• Persona        : %s (%s)\n", persona.Name, persona.TargetLanguage)
	fmt.Printf("• Target Topic   : \"%s\"\n", *topic)
	fmt.Printf("• Master Target  : %s\n", *outputPath)
	if *ttsURL != "" {
		fmt.Printf("• TTS Backend    : %s (Cloud Neural Engine)\n", *ttsURL)
	} else {
		fmt.Println("• TTS Backend    : Local Edge-TTS (en-US-ChristopherNeural)")
	}
	fmt.Println("--------------------------------------------------")

	ctx := context.Background()

	// 2. Script Generation Stage
	fmt.Println("⏳ [1/4] Generating daily introspective script...")
	scriptGen := prompt.NewBuiltinScriptGenerator()
	script, err := scriptGen.Generate(ctx, persona, *topic)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Script generation failed: %v\n", err)
		os.Exit(1)
	}

	// 3. Acoustic Boundary Chunking Stage
	fmt.Println("⏳ [2/4] Partitioning script into acoustic boundary chunks...")
	textChunker := chunker.NewBoundaryChunker(100, 250)
	chunks, err := textChunker.Chunk(script)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Text chunking failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("      -> Successfully formed %d safe acoustic fragments.\n", len(chunks))

	// 4. Voice Synthesis Stage
	fmt.Println("⏳ [3/4] Synthesizing speech across all acoustic chunks...")
	var synth synthesizer.Synthesizer
	if *ttsURL != "" {
		synth = synthesizer.NewHTTPSynthesizer(*ttsURL)
	} else {
		synth = synthesizer.NewEdgeTTSSynthesizer("en-US-ChristopherNeural", "-8%")
	}

	cacheDir := filepath.Join("storage", "cache")
	var chunkPaths []string

	for _, c := range chunks {
		chunkWav := filepath.Join(cacheDir, fmt.Sprintf("chunk_%02d.wav", c.Index))
		fmt.Printf("      -> Synthesizing Chunk [%d/%d] (%d chars)...\n", c.Index+1, len(chunks), len(c.Text))

		if err := synth.SynthesizeChunk(ctx, c, persona, chunkWav); err != nil {
			fmt.Fprintf(os.Stderr, "Synthesis failed for chunk %d: %v\n", c.Index, err)
			os.Exit(1)
		}
		chunkPaths = append(chunkPaths, chunkWav)
	}

	// 5. Audio Post-Processing: Crossfade, Ambient Layering & EBU R128 Mastering
	fmt.Println("⏳ [4/4] Executing 100ms crossfade stitching & EBU R128 mastering...")
	mixer := audio.NewFFmpegMixer()
	if err := mixer.StitchAndMaster(ctx, chunkPaths, *ambientPath, *outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Audio mastering failed: %v\n", err)
		os.Exit(1)
	}

	// 6. Completion
	fmt.Println("==================================================")
	fmt.Println("🎉 SUCCESS! Daily Master Audio Generated!")
	fmt.Println("==================================================")
	fmt.Printf("📁 Playable Master File: %s\n\n", *outputPath)
}

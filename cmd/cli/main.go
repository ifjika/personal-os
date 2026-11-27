package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"personal-os/internal/audio"
	"personal-os/internal/chunker"
	"personal-os/internal/dispatcher"
	"personal-os/internal/domain"
	"personal-os/internal/prompt"
	"personal-os/internal/synthesizer"
)

func main() {
	loadEnvFile(".env")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "generate":
		handleGenerate(os.Args[2:])
	case "version":
		fmt.Println("Personal OS Engine v0.4.2")
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
	fmt.Println("  --topic       Target theme for motivation monologue (optional; auto-generates if omitted)")
	fmt.Println("  --persona     Vocal persona preset (default: elliot)")
	fmt.Println("  --tts-url     Remote/Cloud TTS endpoint URL (e.g. Hugging Face Space)")
	fmt.Println("  --gemini-key  Google Gemini API key for dynamic AI monologues (or GEMINI_API_KEY env)")
	fmt.Println("  --ambient     Path to ambient background audio track (optional)")
	fmt.Println("  --output      Path for the final master MP3 audio output")
	fmt.Println("  --telegram    Dispatch master audio to Telegram Bot upon completion (optional)")
	fmt.Println("  --tg-token    Telegram Bot Token (or TELEGRAM_BOT_TOKEN env)")
	fmt.Println("  --tg-chat-id  Telegram Chat ID (or TELEGRAM_CHAT_ID env)")
}

func handleGenerate(args []string) {
	startTime := time.Now()
	fs := flag.NewFlagSet("generate", flag.ExitOnError)

	topic := fs.String("topic", "", "Motivation monologue theme (optional; dynamic if omitted)")
	personaID := fs.String("persona", "elliot", "Vocal persona identifier")
	ttsURL := fs.String("tts-url", os.Getenv("TTS_URL"), "Cloud or local TTS endpoint URL (defaults to TTS_URL env)")
	ambientPath := fs.String("ambient", "", "Path to ambient background track (optional)")
	geminiKey := fs.String("gemini-key", os.Getenv("GEMINI_API_KEY"), "Google Gemini API key (defaults to GEMINI_API_KEY env)")
	enableTelegram := fs.Bool("telegram", false, "Send master MP3 to Telegram Bot")
	tgToken := fs.String("tg-token", os.Getenv("TELEGRAM_BOT_TOKEN"), "Telegram Bot Token")
	tgChatID := fs.String("tg-chat-id", os.Getenv("TELEGRAM_CHAT_ID"), "Telegram Chat ID")

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
	var scriptGen prompt.ScriptGenerator
	if *geminiKey != "" {
		if *topic != "" {
			fmt.Printf("⏳ [1/4] Generating dynamic script with Gemini AI (Topic: \"%s\")...\n", *topic)
		} else {
			fmt.Println("⏳ [1/4] Generating dynamic script with Gemini AI (Auto-generated novel topic)...")
		}
		scriptGen = prompt.NewGeminiScriptGenerator(*geminiKey)
	} else {
		fmt.Println("⏳ [1/4] Generating script using Builtin template (Set GEMINI_API_KEY for dynamic AI)...")
		scriptGen = prompt.NewBuiltinScriptGenerator()
	}

	script, err := scriptGen.Generate(ctx, persona, *topic)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Script generation failed: %v\n", err)
		os.Exit(1)
	}
	words := len(strings.Fields(script))
	if geminiG, ok := scriptGen.(*prompt.GeminiScriptGenerator); ok && geminiG.LastUsage.TotalTokens > 0 {
		fmt.Printf("      -> Monologue generated: %d words (%d chars) [Tokens: %d in / %d out / %d total].\n",
			words, len(script), geminiG.LastUsage.PromptTokens, geminiG.LastUsage.CandidateTokens, geminiG.LastUsage.TotalTokens)
	} else {
		fmt.Printf("      -> Monologue generated: %d words (%d chars).\n", words, len(script))
	}

	// 3. Acoustic Boundary Chunking Stage
	fmt.Println("⏳ [2/4] Partitioning script into acoustic boundary chunks...")
	textChunker := chunker.NewBoundaryChunker(120, 280)
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
	fmt.Println("⏳ [4/4] Executing equal-power 250ms crossfade stitching & EBU R128 mastering...")
	mixer := audio.NewFFmpegMixer()
	if err := mixer.StitchAndMaster(ctx, chunkPaths, *ambientPath, *outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Audio mastering failed: %v\n", err)
		os.Exit(1)
	}

	// 6. Optional Telegram Bot Dispatch
	var tgSent bool
	if *enableTelegram || (*tgToken != "" && *tgChatID != "") {
		fmt.Println("⏳ [5/5] Dispatching master audio to Telegram Bot...")
		tg := dispatcher.NewTelegramDispatcher(*tgToken, *tgChatID)
		summary := *topic
		if geminiG, ok := scriptGen.(*prompt.GeminiScriptGenerator); ok && geminiG.LastSummary != "" {
			summary = geminiG.LastSummary
		}
		caption := fmt.Sprintf("🎙️ *Personal OS — Motivation Engine*\n\n📌 *Topic:* _%s_\n📝 *Summary:* _%s_\n\n_\"Hello, friend...\"_", *topic, summary)
		title := fmt.Sprintf("Motivation: %s", *topic)
		if err := tg.SendAudio(ctx, *outputPath, caption, title, persona.Name); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️ Telegram dispatch failed: %v\n", err)
		} else {
			tgSent = true
			fmt.Printf("      -> Successfully delivered audio to Telegram (Chat ID: %s)!\n", *tgChatID)
		}
	}

	// 7. Completion
	elapsed := time.Since(startTime)
	fmt.Println("==================================================")
	fmt.Println("🎉 SUCCESS! Daily Master Audio Generated!")
	fmt.Println("==================================================")
	fmt.Printf("📁 Playable Master File : %s\n", *outputPath)
	fmt.Printf("⏱️  Total Process Time  : %s\n", elapsed.Round(100*time.Millisecond))
	if geminiG, ok := scriptGen.(*prompt.GeminiScriptGenerator); ok && geminiG.LastUsage.TotalTokens > 0 {
		fmt.Printf("🪙  Gemini Tokens Spent : %d tokens (%d prompt / %d output)\n",
			geminiG.LastUsage.TotalTokens, geminiG.LastUsage.PromptTokens, geminiG.LastUsage.CandidateTokens)
	} else {
		fmt.Println("🪙  Gemini Tokens Spent : N/A (Offline Builtin template used)")
	}
	if tgSent {
		fmt.Printf("📲 Telegram Delivery    : Sent to Chat %s\n", *tgChatID)
	}
	fmt.Println("==================================================")
}

// loadEnvFile reads key=value pairs from a .env file and sets environment variables if not already present.
func loadEnvFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return // .env is optional
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)
			if os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
	}
}

# Design System & UI Specification
## Personal OS: Terminal Noir / Cyber-Minimalist Aesthetic

---

## 1. Design Philosophy & Sensory Tone

The visual design system of **Personal OS** mirrors its auditory thesis: dark, focused, tactile, and stripped of distracting ornamentation. Drawing inspiration from terminal emulators, cyber-investigation suites (e.g., *Mr. Robot* command line interfaces), and high-end audio mastering hardware, the design adheres to the **Terminal Noir** archetype:

- **Pitch-Black Canvases:** Deep, absolute darkness reduces OLED battery consumption and eye strain during late-night deep work sessions.
- **Phosphor Accent Cues:** High-contrast tactical accents in phosphor cyan (`#00F0FF`), terminal phosphor green (`#00FF66`), and warm amber (`#FFB800`) are used sparingly to communicate machine state and audio levels.
- **Acoustic Tactility:** Buttons, sliders, and audio scrubbers provide immediate micro-interactions that feel mechanical and deliberate.
- **Information Density:** Prioritizes dense, high-signal information displays (audio waveforms, terminal logs, spectrum levels) over superficial empty space.

---

## 2. Design Tokens

### 2.1 Color Palette

```text
BACKGROUND / NEUTRALS
┌─────────────────┬─────────────────┬─────────────────┬─────────────────┐
│ Jet Black       │ Obsidian Void   │ Slate Surface   │ Muted Border    │
│ #090A0F         │ #0F1117         │ #181B26         │ #262B3D         │
│ (Base App Canvas│ (Card Surfaces) │ (Active States) │ (Hairline Rules)│
└─────────────────┴─────────────────┴─────────────────┴─────────────────┘

TACTICAL ACCENTS
┌─────────────────┬─────────────────┬─────────────────┬─────────────────┐
│ Phosphor Cyan   │ Terminal Green  │ Amber Alert     │ Crimson Error   │
│ #00F0FF         │ #00FF66         │ #FFB800         │ #FF3366         │
│ (Active Focus)  │ (Job Success)   │ (Processing)    │ (GPU Crash/OOM) │
└─────────────────┴─────────────────┴─────────────────┴─────────────────┘

TYPOGRAPHY SCALES
┌─────────────────┬─────────────────┬─────────────────┬─────────────────┐
│ Primary Text    │ Secondary Text  │ Dim Subtext     │ Disabled Code   │
│ #F1F5F9         │ #94A3B8         │ #64748B         │ #334155         │
└─────────────────┴─────────────────┴─────────────────┴─────────────────┘
```

#### Tailwind Configuration Tokens (`tailwind.config.js` extension)
```javascript
module.exports = {
  theme: {
    extend: {
      colors: {
        canvas: {
          base: '#090A0F',
          surface: '#0F1117',
          elevated: '#181B26',
          border: '#262B3D',
        },
        brand: {
          cyan: '#00F0FF',
          green: '#00FF66',
          amber: '#FFB800',
          crimson: '#FF3366',
        },
        syntax: {
          primary: '#F1F5F9',
          muted: '#94A3B8',
          dim: '#64748B',
        }
      }
    }
  }
}
```

### 2.2 Typography Hierarchy

We standardize on a dual-font structure:
1. **Monospace Engine:** `JetBrains Mono`, `Fira Code`, or `Geist Mono` for all data displays, acoustic tags, terminal output, and technical metrics.
2. **System Interface:** `Inter` or `Geist Sans` for structural UI labels, modal headers, and general prose.

| Style Role | Font Family | Size | Weight | Line Height | Tracking |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Display H1** | `Geist Mono` | 24px (1.5rem) | 700 (Bold) | 32px | -0.02em |
| **Section H2** | `Geist Mono` | 18px (1.125rem) | 600 (Semibold) | 26px | -0.01em |
| **Metric Value** | `JetBrains Mono` | 20px (1.25rem) | 700 (Bold) | 28px | 0.05em |
| **Body Text** | `Inter` | 14px (0.875rem) | 400 (Regular) | 22px | 0.00em |
| **Terminal Log** | `JetBrains Mono` | 12px (0.75rem) | 400 (Regular) | 18px | 0.02em |
| **Acoustic Tag** | `JetBrains Mono` | 11px (0.6875rem) | 500 (Medium) | 14px | 0.08em uppercase |

### 2.3 Spacing & Layout Grid
- **Base Unit:** 4px geometric scaling (`4px`, `8px`, `12px`, `16px`, `24px`, `32px`, `48px`).
- **Borders & Dividers:** 1px hairline borders (`border border-canvas-border`) with zero blur drop-shadows. Shadows are replaced with subtle 1px border highlights (`box-shadow: 0 0 12px rgba(0, 240, 255, 0.15)` on focus).
- **Corner Radii:** Strict, minimalist geometry: `rounded-none` or subtle `rounded-sm` (2px) to `rounded-md` (4px). Rounded "pill" designs are forbidden except for status indicator dots.

---

## 3. Core UI Component Specifications

### 3.1 Audio Waveform Scrubber & Visualizer
- **Visual Presentation:** Renders a real-time canvas-based stereo audio waveform.
- **States:**
  - *Inactive/Unloaded:* Muted grid lines (`#262B3D`).
  - *Playing:* Phosphor cyan fill (`#00F0FF`) with ambient glow on the playhead.
  - *Buffering / Inferencing:* Amber pulse (`#FFB800`) across unrendered segments.
- **Interaction:** Drag or click to scrub; microsecond-accurate timecode displayed on hover in `JetBrains Mono`.

### 3.2 Acoustic Directive Chip (`TagBadge`)
- Used in the Script Studio to highlight special acoustic directions:
  - `[whisper]` $\to$ Cyan border, faint cyan translucent fill (`rgba(0,240,255,0.08)`).
  - `[sigh]` $\to$ Dim slate border, muted grey text.
  - `[pause:2s]` $\to$ Amber outline, indicating a temporal silence insertion.

### 3.3 Ambient Soundscape Mixer Sliders
- **Controls:** Multi-track faders for:
  1. Primary Vocal (Elliot / Batman)
  2. Brown Noise Generator
  3. Steady Rain / Window Droplets
  4. Room Tone / Vinyl Hiss
- **Behaviors:** Independent decibel meters ($-60 \text{ dB}$ to $+6 \text{ dB}$), mute toggle, and solo switch. Smooth logarithmic attenuation curves.

### 3.4 Terminal Log Streamer
- Real-time Server-Sent Events (SSE) log console displaying backend orchestration actions:
  ```text
  [20:45:12.310] [ORCHESTRATOR] [INFO] Script generated via DeepSeek-V3 (1,248 tokens)
  [20:45:12.390] [CHUNKER]      [DEBUG] Extracted 14 acoustic fragments
  [20:45:13.100] [GPU_WORKER]   [INFO] Dispatched chunk 01/14 to F5-TTS (VRAM: 4,120 MB)
  [20:45:14.280] [FFMPEG]       [INFO] Stitching chunk 01-02 with 100ms crossfade
  ```

---

## 4. UX Rules & Interaction Principles

1. **Zero Autoplay Without Explicit Consent:** High-intensity audio must never auto-trigger upon application load. All playback requires a deliberate user keystroke (`Space`) or mouse click.
2. **Keyboard-First Navigation:**
   - `Space`: Play / Pause active master audio.
   - `J` / `L`: Seek backward / forward by 5 seconds.
   - `M`: Toggle master mute.
   - `Ctrl + Enter` / `Cmd + Enter`: Trigger generation pipeline from the Script Editor.
   - `Escape`: Dismiss active modal or stop streaming preview.
3. **Accessibility Standards:**
   - High contrast compliance (WCAG 2.1 AAA) for all textual content against dark canvases.
   - Visible keyboard focus rings (`outline: 1px solid #00F0FF`).
   - Screen-reader aria attributes on all multi-track mixer sliders and audio scrubbers.
4. **State Machine Transparency:** The UI must never freeze or present an indeterminate spinner. Every background task must show clear phase transitions: `PROMPTING` $\to$ `CHUNKING` $\to$ `SYNTHESIZING [X/Y]` $\to$ `MASTERING` $\to$ `READY`.

# Testing Guide & Quality Assurance Standards
## Personal OS: AI Audio Motivation Engine

---

## 1. Testing Philosophy & Standards

Personal OS handles sensitive multimedia transformations and concurrent pipeline orchestrations. All tests adhere to the **Arrange-Act-Assert (AAA)** pattern with strict determinism:

- **Zero Network Flakiness:** Unit tests must never make outbound internet calls to external LLMs or third-party APIs. All external network boundaries must be mocked.
- **Race Condition Immunity:** All Go tests must pass the Go race detector (`go test -race ./...`).
- **Acoustic Integrity:** Audio generation tests must programmatically verify audio headers, sample rates, and RMS dynamic range rather than simply asserting non-empty file sizes.

---

## 2. Unit Testing in Golang (Table-Driven Tests)

### 2.1 Boundary Chunker Test Example (`internal/chunker/chunker_test.go`)

```go
package chunker_test

import (
	"testing"
	"personal-os/apps/orchestrator/internal/chunker"
)

func TestChunkScript_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		inputScript   string
		minSize       int
		maxSize       int
		expectedCount int
		hasTag        string
	}{
		{
			name:          "Standard paragraph with acoustic tags",
			inputScript:   "[whisper] The world is loud right now. Breathe. Let the noise fade into the background. [pause] Focus on the machine.",
			minSize:       20,
			maxSize:       100,
			expectedCount: 2,
			hasTag:        "whisper",
		},
		{
			name:          "Single short sentence",
			inputScript:   "Stay disciplined.",
			minSize:       50,
			maxSize:       150,
			expectedCount: 1,
			hasTag:        "",
		},
		{
			name:          "Preserve bracketed pause without slicing",
			inputScript:   "Line one of code. [pause:3s] Line two of code. Let the algorithm breathe.",
			minSize:       20,
			maxSize:       80,
			expectedCount: 2,
			hasTag:        "pause:3s",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange & Act
			c := chunker.NewBoundaryChunker(tc.minSize, tc.maxSize)
			chunks, err := c.Chunk(tc.inputScript)

			// Assert
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(chunks) != tc.expectedCount {
				t.Errorf("expected %d chunks, got %d", tc.expectedCount, len(chunks))
			}
			if tc.hasTag != "" && chunks[0].AcousticTag != tc.hasTag {
				t.Errorf("expected acoustic tag %q, got %q", tc.hasTag, chunks[0].AcousticTag)
			}
		})
	}
}
```

### 2.2 Testing FFmpeg Command Construction
Unit tests verify that dynamic `filter_complex` arguments are assembled deterministically without invoking the host binary:

```go
func TestBuildFFmpegStitchArgs(t *testing.T) {
    chunkFiles := []string{"chunk_0.wav", "chunk_1.wav", "chunk_2.wav"}
    args := audio.BuildStitchCommand(chunkFiles, 100, "master.wav")
    
    // Assert correct number of inputs and acrossfade count
    expectedFilter := "[0:a][1:a]acrossfade=d=0.1:c1=tri:c2=tri[a01];[a01][2:a]acrossfade=d=0.1:c1=tri:c2=tri[out]"
    if !containsSubslice(args, expectedFilter) {
        t.Errorf("filter graph mismatch. got: %v", args)
    }
}
```

---

## 3. Integration Testing & Mocking Standards

### 3.1 Mocking the Python Inference Microservice
We use Go's native `net/http/httptest` to simulate the FastAPI worker, returning synthetic uncompressed WAV headers to validate pipeline flow:

```go
func setupMockPythonWorker(t *testing.T) *httptest.Server {
    return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.URL.Path {
        case "/v1/health":
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusOK)
            w.Write([]byte(`{"status":"HEALTHY","gpu":{"free_vram_mb":12000}}`))
        case "/v1/synthesize":
            w.Header().Set("Content-Type", "audio/wav")
            w.WriteHeader(http.StatusOK)
            // Write 1 second of clean synthetic WAV header + silence
            w.Write(generateSyntheticWavHeader(48000, 1))
        default:
            w.WriteHeader(http.StatusNotFound)
        }
    }))
}
```

### 3.2 In-Memory Database Tests
All repository integration tests execute against an in-memory SQLite database (`:memory:`) to ensure lightning-fast setup and teardown:

```go
func setupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
    if err != nil {
        t.Fatalf("failed to open in-memory sqlite: %v", err)
    }
    // Run migrations
    repository.RunMigrations(db)
    return db
}
```

---

## 4. Python Microservice Testing (`pytest`)

Audio testing within the Python worker verifies neural network output validity and tensor stability:

```python
import pytest
from httpx import AsyncClient
from src.main import app

@pytest.mark.asyncio
async def test_health_check():
    async with AsyncClient(app=app, base_url="http://test") as client:
        response = await client.get("/v1/health")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "HEALTHY"
        assert "gpu" in data

@pytest.mark.asyncio
async def test_synthesize_validation_error():
    async with AsyncClient(app=app, base_url="http://test") as client:
        # Submit empty text
        response = await client.post("/v1/synthesize", json={"text": ""})
        assert response.status_code == 422
```

---

## 5. CI/CD Pipeline Automation (GitHub Actions)

File: `.github/workflows/ci.yml`

```yaml
name: Test Suite & Security Linting

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test-orchestrator:
    name: Golang Orchestrator Tests
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Install FFmpeg
        run: sudo apt-get update && sudo apt-get install -y ffmpeg
      - name: Run Go Linters
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest
          working-directory: apps/orchestrator
      - name: Execute Go Tests with Race Detector
        run: |
          cd apps/orchestrator
          go test -race -v -coverprofile=coverage.out ./...

  test-python-worker:
    name: Python Worker Tests
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.10'
      - name: Install Dependencies
        run: |
          cd apps/inference-worker
          pip install -r requirements-test.txt
      - name: Run Pytest
        run: |
          cd apps/inference-worker
          pytest -v tests/
```

---

## 6. Audio Artifact Sanity Checklist

Before accepting a generated master audio track into production:
1. **Header Check:** Verify valid RIFF/WAVE or MP3 ID3 header.
2. **True Peak Analysis:** Ensure True Peak does not clip ($>-1.0\text{ dBFS}$).
3. **Integrated Loudness:** Must fall strictly within $-16.0 \pm 1.0\text{ LUFS}$.
4. **Duration Check:** Total duration must match the sum of component chunk durations within $\pm 2.0\%$.

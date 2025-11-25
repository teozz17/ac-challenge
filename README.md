# AC Challenge Solution

## How to Run

1.  **Start Infrastructure**:
    ```bash
    docker-compose up -d
    ```
    This spins up MongoDB and Jaeger (for tracing).

2.  **Environment Variables**:
    Make sure you have these exported in your shell:
    ```bash
    export OPENAI_API_KEY="sk-..."
    export WEATHER_API_KEY="your_key_here" # Check the email I sent for this one!
    ```

3.  **Run the Server**:
    ```bash
    make run
    ```
    The server will start on port 8080.

## What I Did

### Task 1: Fix Title Generation
The title wasn't being saved because the system instruction was being overwritten incorrectly. I fixed that logic, and also made the title and reply generation run in parallel. Now the response is snappier.

### Task 2: Weather API
I integrated **WeatherAPI.com** to get real weather data.
- Supports current weather.
- Supports forecasts (up to 14 days).
- You can ask for "weather in Barcelona next Monday" or "hourly forecast for tomorrow".

### Task 3: Refactor Tools
The tool logic in `assistant.go` was getting messy, so I refactored it out.
- Created a `internal/chat/tools` package.
- Implemented a **Registry** pattern to manage tools easily.
- **Bonus**: Added a `get_time_in_zone` tool. You can ask "What time is it in Tokyo?" and it works.
### Task 4: Automated Tests
I added a mix of unit and integration tests:
- **Server Tests**: Uses a `MockAssistant` to test the API endpoints without hitting OpenAI (fast & free).
- **Assistant Tests (Bonus)**: Added an integration test that actually hits OpenAI to verify the prompts work. It skips automatically if you don't have the API key set.

### Task 5: Instrumentation (Observability)
I instrumented the server with **OpenTelemetry**.
- **Metrics**: Basic metrics (request count, duration) are printed to stdout (kept it simple as requested).
- **Tracing (Bonus)**: I added **Jaeger** to the docker-compose.
    - Go to `http://localhost:16686` to see the traces.
    - I added manual spans so you can see exactly how long "GenerateTitle" vs "GenerateReply" takes in the waterfall view.

---
Let me know if you have any questions!
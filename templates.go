package main

// htmlTemplate contains the HTML template for the flag quiz game
const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Flag Quiz Game</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 900px;
            margin: 20px auto;
            padding: 20px;
            background-color: #f0f8ff;
        }
        .container {
            background-color: white;
            padding: 30px;
            border-radius: 15px;
            box-shadow: 0 4px 15px rgba(0,0,0,0.1);
        }
        h1 {
            color: #2c3e50;
            text-align: center;
            margin-bottom: 10px;
        }
        .subtitle {
            text-align: center;
            color: #7f8c8d;
            margin-bottom: 30px;
            font-style: italic;
        }
        .game-area {
            text-align: center;
            margin: 30px 0;
        }
        .flag-container {
            display: inline-block;
            border: 3px solid #34495e;
            border-radius: 10px;
            overflow: hidden;
            margin: 20px 0;
            box-shadow: 0 2px 10px rgba(0,0,0,0.2);
        }
        .flag-image {
            display: block;
            width: 300px;
            height: auto;
            max-height: 200%;
        }
        .question {
            font-size: 24px;
            color: #2c3e50;
            margin: 20px 0;
            font-weight: bold;
        }
        .country-name {
            font-size: 20px;
            color: #e74c3c;
            margin: 15px 0;
            font-weight: bold;
        }
        .buttons {
            margin: 30px 0;
        }
        .btn {
            font-size: 18px;
            padding: 15px 30px;
            margin: 10px;
            border: none;
            border-radius: 8px;
            cursor: pointer;
            font-weight: bold;
            transition: all 0.3s ease;
        }
        .btn-correct {
            background-color: #27ae60;
            color: white;
        }
        .btn-correct:hover {
            background-color: #229954;
        }
        .btn-incorrect {
            background-color: #e74c3c;
            color: white;
        }
        .btn-incorrect:hover {
            background-color: #c0392b;
        }
        .btn-new {
            background-color: #3498db;
            color: white;
        }
        .btn-new:hover {
            background-color: #2980b9;
        }
        .result {
            font-size: 20px;
            font-weight: bold;
            margin: 20px 0;
            padding: 15px;
            border-radius: 8px;
        }
        .result.correct {
            background-color: #d5f4e6;
            color: #27ae60;
            border: 2px solid #27ae60;
        }
        .result.incorrect {
            background-color: #fadbd8;
            color: #e74c3c;
            border: 2px solid #e74c3c;
        }
        .score {
            background-color: #ebf3fd;
            padding: 15px;
            border-radius: 8px;
            margin: 20px 0;
            text-align: center;
        }
        .score h3 {
            margin: 0 0 10px 0;
            color: #2c3e50;
        }
        .stats {
            display: flex;
            justify-content: space-around;
            flex-wrap: wrap;
        }
        .stat {
            margin: 5px;
        }
        .stat-value {
            font-size: 24px;
            font-weight: bold;
            color: #3498db;
        }
        .stat-label {
            font-size: 14px;
            color: #7f8c8d;
        }
        .flag-comparison {
            display: flex;
            justify-content: center;
            gap: 20px;
            margin: 20px 0;
            flex-wrap: wrap;
        }
        .flag-box {
            text-align: center;
            padding: 10px;
            border-radius: 8px;
            background-color: #f8f9fa;
            border: 2px solid #dee2e6;
            min-width: 120px;
        }
        .flag-box h4 {
            margin: 0 0 10px 0;
            font-size: 14px;
            color: #495057;
        }
        .flag-thumbnail {
            display: block;
            max-width: 100%;
            height: auto;
            border: 2px solid #333;
            border-radius: 4px;
            object-fit: cover;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🏴 Flag Quiz Game</h1>
        <div class="subtitle">Can you spot the fake flags?</div>
        
        <div class="score">
            <h3>Your Score</h3>
            <div class="stats">
                <div class="stat">
                    <div class="stat-value" id="correct">{{.Correct}}</div>
                    <div class="stat-label">Correct</div>
                </div>
                <div class="stat">
                    <div class="stat-value" id="incorrect">{{.Incorrect}}</div>
                    <div class="stat-label">Incorrect</div>
                </div>
                <div class="stat">
                    <div class="stat-value" id="total">{{.Total}}</div>
                    <div class="stat-label">Total</div>
                </div>
                <div class="stat">
                    <div class="stat-value" id="percentage">{{.Percentage}}%</div>
                    <div class="stat-label">Accuracy</div>
                </div>
            </div>
        </div>

        <div class="game-area">
            {{if .FlagData}}
                <div class="question">Is this the correct flag?</div>
                <div class="country-name">{{.CountryName}}</div>
                <div class="flag-container">
                    <img src="data:image/png;base64,{{.FlagData}}" alt="Flag" class="flag-image">
                </div>
                
                {{if .ShowResult}}
                    <div class="result {{if .ResultCorrect}}correct{{else}}incorrect{{end}}">
                        {{if .ResultCorrect}}
                            ✓ Correct! {{.ResultMessage}}
                        {{else}}
                            ✗ Wrong! {{.ResultMessage}}
                            {{if and .OriginalFlag .ModifiedFlag (not .IsCorrect)}}
                                <div class="flag-comparison">
                                    <div class="flag-box">
                                        <h4>Correct Flag</h4>
                                        <img src="data:image/png;base64,{{.OriginalFlag}}" alt="Correct Flag" class="flag-thumbnail">
                                    </div>
                                </div>
                            {{end}}
                        {{end}}
                    </div>
                    <button class="btn btn-new" onclick="location.href='/new'">Next Flag</button>
                {{else}}
                    <div class="buttons">
                        <button class="btn btn-correct" onclick="location.href='/guess?answer=correct'">Correct Flag</button>
                        <button class="btn btn-incorrect" onclick="location.href='/guess?answer=incorrect'">Fake Flag</button>
                    </div>
                {{end}}
            {{else}}
                <div class="question">Welcome to the Flag Quiz!</div>
                <p>Test your knowledge of world flags. Some flags will be correct, others will have subtle errors.</p>
                <button class="btn btn-new" onclick="location.href='/new'">Start Game</button>
            {{end}}
        </div>
    </div>
</body>
</html>
`

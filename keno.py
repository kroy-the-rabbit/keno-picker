from flask import Flask, request
import requests
import random
import re
from time import sleep

app = Flask(__name__)

def fetch_seed():
    response = requests.get("https://rand.kroy.io")
    response.raise_for_status()
    return response.text

def generate_unique_random_numbers(seed, count):
    random.seed(seed)
    return random.sample(range(1, 81), count)

# Flip a coin `times` times.  Make it hacky so it's quick
def flip_coin(times):
    results = [random.getrandbits(1) for _ in range(times)]
    heads_count = sum(results)
    # Heads == true, tails == false
    return heads_count > (times - heads_count)

@app.route('/', methods=['GET'])
def random_numbers():
    count = request.args.get('count', default=4, type=int)
    count = max(1, min(count, 20))
    seed = fetch_seed()
    numbers = generate_unique_random_numbers(seed, count)

    sleep(count * .01)
    seed = fetch_seed()
    threespot = generate_unique_random_numbers(seed, 3)

    sleep(count * .001)
    seed = fetch_seed()
    alternate = generate_unique_random_numbers(seed, 4)
    
    heads_wins = flip_coin(574673)
    coin_flip = "A" if heads_wins else "B" 
    
    # dark mode because fuckoff
    html = '''
    <html>
    <head>
        <style>
            body {{
                background-color: #333;
                color: #fff;
                font-family: Arial, sans-serif;
                margin: 40px;
            }}
            hr {{
                border: 1px solid #555;
            }}
        </style>
    </head>
    <body>
        <h2>Keno Picker</h2>
        Seed: {seed}<br><hr><br>Result A:<br> {number_list}
        <br><hr><br>Three Spot:<br> {threespot_list}
        <br><hr><br>Result B:<br> {alternate_list}
        <br><hr><br>Which result to pick: {result}
    </body>
    </html>
    '''.format(seed=seed, alternate_list='<br>'.join(map(str, alternate)), threespot_list='<br>'.join(map(str, threespot)), number_list='<br>'.join(map(str, numbers)), result=coin_flip)

    
    return html

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)


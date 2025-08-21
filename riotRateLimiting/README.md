# Riot Rate Limiting
Uses a clock to limit requests to the Riot Games API in accordance to the limits set for personal API keys  
Each availble call is represented by an element in a queue as it is the only redis datastructure with a blocking command clearly documented on the redis website that I could find

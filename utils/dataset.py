import csv
import random
from datetime import datetime, timedelta

# dates between
def generate_random_date(start_year=2023, end_year=2023):
    start = datetime(start_year, 1, 1)
    end = datetime(end_year + 1, 1, 1)
    return start + (end - start) * random.random()

# event range
event_descriptions = [
    "Login", "View Product", "Add to Cart", "Search", "Logout",
    "Register", "Place Order", "Checkout", "Payment", "Confirmation"
]

# settings for generations
num_sessions = 1000000  # numb sessions
events_per_session = 5  # max event on session
time_between_events = timedelta(minutes=5)  # min time between events

# Открываем файл для записи
with open('largest_dataset.csv', 'w', newline='') as file:
    writer = csv.writer(file)
    writer.writerow(["SessionID", "Timestamp", "Description"])  # headers
    
    for session_id in range(1, num_sessions + 1):  
        session_start_time = generate_random_date()  # gen
        
        num_events = random.randint(1, events_per_session)
        events = random.choices(event_descriptions, k=num_events) 
        
        for i, event in enumerate(events):
            timestamp = session_start_time + i * time_between_events
            timestamp_str = timestamp.strftime("%Y-%m-%dT%H:%M:%SZ")  # time ISO 8601
            
            # to file
            writer.writerow([session_id, timestamp_str, event])
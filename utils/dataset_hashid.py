import csv
import random
from datetime import datetime, timedelta
import hashlib  

def generate_random_date(start_year=2023, end_year=2023):
    start = datetime(start_year, 1, 1)
    end = datetime(end_year + 1, 1, 1)
    return start + (end - start) * random.random()

event_descriptions = [
    "Login", "View Product", "Add to Cart", "Search", "Logout",
    "Register", "Place Order", "Checkout", "Payment", "Confirmation"
]

num_sessions = 3000000  
events_per_session = 10  
time_between_events = timedelta(minutes=5) 

# для генерации псевдохешированного SessionID
def generate_hashed_session_id(session_id):
    hash_object = hashlib.md5(str(session_id).encode())
    return hash_object.hexdigest()  

with open('datasets/largest_dataset5.csv', 'w', newline='') as file:
    writer = csv.writer(file)
    writer.writerow(["SessionID", "Timestamp", "Description"])  
    
    for session_id in range(1, num_sessions + 1):  
        
        hashed_session_id = generate_hashed_session_id(session_id)
        session_start_time = generate_random_date()
  
        num_events = random.randint(8, events_per_session)
        events = random.choices(event_descriptions, k=num_events) 
        
        for i, event in enumerate(events):
            timestamp = session_start_time + i * time_between_events
            timestamp_str = timestamp.strftime("%Y-%m-%dT%H:%M:%SZ")  # ISO 8601
            
            writer.writerow([hashed_session_id, timestamp_str, event])
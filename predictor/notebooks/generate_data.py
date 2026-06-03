import pandas as pd
import numpy as np
from datetime import datetime, timedelta
import random

random.seed(42)
np.random.seed(42)

ZONES = {
    "koramangala":    {"lat": 12.9352, "lng": 77.6245, "base_demand": 18},
    "indiranagar":    {"lat": 12.9784, "lng": 77.6408, "base_demand": 15},
    "whitefield":     {"lat": 12.9698, "lng": 77.7499, "base_demand": 12},
    "marathahalli":   {"lat": 12.9591, "lng": 77.6974, "base_demand": 14},
    "hsr_layout":     {"lat": 12.9116, "lng": 77.6389, "base_demand": 13},
    "jp_nagar":       {"lat": 12.9102, "lng": 77.5856, "base_demand": 11},
    "electronic_city":{"lat": 12.8399, "lng": 77.6770, "base_demand": 10},
    "hebbal":         {"lat": 13.0353, "lng": 77.5972, "base_demand": 9},
}

def generate_dataset(days=90):
    records = []
    
    start_date = datetime(2024, 1, 1)
    
    for day in range(days):
        current_date = start_date + timedelta(days=day)
        day_of_week = current_date.weekday()
        is_weekend = 1 if day_of_week >= 5 else 0
        is_raining = 1 if random.random() < 0.2 else 0
        
        for hour in range(24):
            for zone_name, zone_info in ZONES.items():
                base = zone_info["base_demand"]
                
                if 11 <= hour <= 13:
                    time_factor = 1.6
                elif 19 <= hour <= 21:
                    time_factor = 1.9
                elif 7 <= hour <= 9:
                    time_factor = 1.2
                elif 0 <= hour <= 5:
                    time_factor = 0.2
                else:
                    time_factor = 0.8
                
                weekend_factor = 1.3 if is_weekend else 1.0
                rain_factor = 1.4 if is_raining else 1.0
                
                prev_hour_orders = max(0, int(
                    base * time_factor * weekend_factor * rain_factor * 
                    random.uniform(0.8, 1.2)
                ))
                
                actual_orders = max(0, int(
                    base * time_factor * weekend_factor * rain_factor *
                    random.uniform(0.85, 1.15)
                ))
                
                noise = np.random.normal(0, 1.5)
                actual_orders = max(0, int(actual_orders + noise))
                
                records.append({
                    "zone_id":           zone_name,
                    "hour_of_day":       hour,
                    "day_of_week":       day_of_week,
                    "is_weekend":        is_weekend,
                    "is_raining":        is_raining,
                    "base_zone_demand":  base,
                    "prev_hour_orders":  prev_hour_orders,
                    "actual_orders":     actual_orders,
                })
    
    return pd.DataFrame(records)

if __name__ == "__main__":
    print("Generating Bangalore delivery dataset...")
    df = generate_dataset(days=90)
    
    print(f"Dataset shape: {df.shape}")
    print(f"\nSample rows:")
    print(df.head(10).to_string())
    
    print(f"\nDemand statistics:")
    print(df["actual_orders"].describe())
    
    print(f"\nAverage orders by hour:")
    hourly = df.groupby("hour_of_day")["actual_orders"].mean().round(1)
    for hour, avg in hourly.items():
        bar = "█" * int(avg)
        print(f"  {hour:02d}:00  {avg:5.1f}  {bar}")
    
    df.to_csv("notebooks/bangalore_delivery_data.csv", index=False)
    print(f"\nSaved to notebooks/bangalore_delivery_data.csv")
    print(f"Total records: {len(df):,}")
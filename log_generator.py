import random
import time
from datetime import datetime
import os

# Log levels and their corresponding messages
LOG_LEVELS = ['INFO', 'WARNING', 'ERROR', 'DEBUG']
SAMPLE_MESSAGES = [
    'User logged in successfully',
    'Database connection established',
    'Failed to process request',
    'Cache cleared',
    'Memory usage high',
    'Network timeout occurred',
    'File processing completed',
    'Invalid input received',
    'System update initiated',
    'Configuration loaded'
]

def generate_log_entry():
    """Generate a random log entry with timestamp, level, and message."""
    timestamp = datetime.now().strftime('%Y-%m-%d %H:%M:%S')
    level = random.choice(LOG_LEVELS)
    message = random.choice(SAMPLE_MESSAGES)
    return f'[{timestamp}] {level}: {message}'

def write_logs(filename='app.txt', num_entries=1000, delay=1):
    """Write random log entries to a file.
    
    Args:
        filename (str): Name of the log file
        num_entries (int): Number of log entries to generate
        delay (int): Delay between entries in seconds
    """
    print(f'Starting to write {num_entries} log entries to {filename}')
    
    for i in range(num_entries):
        log_entry = generate_log_entry()
        
        # Append the log entry to the file
        with open(filename, 'a') as f:
            f.write(log_entry + '\n')
        
        print(f'Written log entry {i+1}/{num_entries}')
        time.sleep(delay)

if __name__ == '__main__':
    # Create logs with 10 entries and 1 second delay between each
    write_logs()
    print('Log generation completed!') 
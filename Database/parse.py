# -*- coding: utf-8 -*-
"""
Created on Sun Dec  5 08:55:24 2021

@author: Libo
"""

from datetime import datetime
import time
import sqlite3
import subprocess
import re
from os import listdir, rename
from os.path import exists

remote_config_path = '../remote_config.txt'
remote_logs_folder = './RemoteLogs/'

def process_text_file(path, con):
    with open(path, 'r') as f:
        lines = f.readlines()
    entries_added = 0
    for line in lines:
        line = re.sub(r'[\[\]|"\n]', '', line)
        parts = line.split(': ')
        all_parts = []
        data = []
        for part in parts:
            if "; " in part:
                all_parts.extend(part.split("; "))
            else:
                all_parts.append(part)
        for index, part in enumerate(all_parts):
            if index == 0:
                timestamp = part
            elif index == 1:
                location = part
            elif index % 2 == 0:
                name = part
            else:
                if part == "TimedOut":
                    data.append((timestamp, location, name))
        
        cur = con.cursor()
        for timestamp, location, name in data:
            #2021-12-05 10:30:41.167
            hour_minute = datetime.strptime(timestamp, '%Y-%m-%d %H:%M:%S.%f').strftime('%H:%M')
            try:
                cur.execute(f"INSERT INTO timeout_data (name, location, timestamp, hour_minute) VALUES ('{name}','{location}','{timestamp}', '{hour_minute}')")
                entries_added += 1
            except sqlite3.IntegrityError as e:
                print(e)
        # Save (commit) the changes
        con.commit()
    
    base_file_name = path.split('.txt')[0]
    if remote_logs_folder in base_file_name:
        base_file_name = base_file_name.replace(remote_logs_folder, '')
        new_path =  f'./Processed/remote-{base_file_name}.txt'
    else:
        new_path =  f'./Processed/{base_file_name}.txt'
    rename(path, new_path)
    print(f'finished processing {path} and moving to {new_path}. Added {entries_added} new entries')
            
def get_text_files(remote_config):
    curr_files = listdir()
    if remote_config:
        subprocess.Popen(f"rsync --remove-source-files -r {remote_config}:/root/logs/ ./RemoteLogs/", shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE).communicate()
    result = [f for f in curr_files if 'txt' in f]
    result.extend([remote_logs_folder + p for p in listdir(remote_logs_folder)])
    return result

if __name__ == "__main__":
    remote_config_file = exists(remote_config_path)
    with open(remote_config_path, "r") as f:
        lines = f.readlines()
        remote_config = lines[0]
    i = 0
    while True:
        curr_time = datetime.now().strftime('%Y-%m-%d %H:%M:%S')
        text_files = get_text_files(remote_config)
        if len(text_files):
            con = sqlite3.connect('data.db')
            print(f"{curr_time}: Processing", text_files)
            for text_file in text_files:
                process_text_file(text_file, con)
            con.close()
        else:
            print(f"{curr_time}: found no new log files")
        time.sleep(60)
        i += 1
        if i % 60 == 0:
            print("uploading db to remote backup")
            subprocess.Popen(f"scp data.db {remote_config}:/root/", shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE).communicate()
            

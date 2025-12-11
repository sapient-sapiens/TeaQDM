import time
import sys
import os

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from python.wrapper import TeaQDM

# Nested Loop Test

epochs = TeaQDM(range(50), desc="Epochs")
for i in epochs:
    batches = TeaQDM(range(50), desc="Batches", parent=epochs)
    for j in batches:
        time.sleep(0.01)  
    mini = TeaQDM(range(3), desc="Batches", parent=epochs)
    for j in mini:
        batches = TeaQDM(range(30), desc="Batches", parent=epochs)
        for j in batches:
            time.sleep(0.01) 
        time.sleep(0.01)  
    
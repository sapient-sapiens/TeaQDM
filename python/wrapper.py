import ctypes
import os
import threading

_so_path = os.path.join(os.path.dirname(os.path.dirname(__file__)), "teaqdm.so")
lib = ctypes.CDLL(os.path.abspath(_so_path))

lib.AddBar.argtypes = [ctypes.c_longlong, ctypes.c_longlong, ctypes.c_char_p]
lib.AddBarWithParent.argtypes = [ctypes.c_longlong, ctypes.c_longlong, ctypes.c_char_p, ctypes.c_longlong]
lib.UpdateBar.argtypes = [ctypes.c_longlong, ctypes.c_longlong]
lib.RemoveBar.argtypes = [ctypes.c_longlong]
lib.StartEngine.argtypes = []

_engine_started = False
_engine_lock = threading.Lock()

def _ensure_engine_started():
    global _engine_started
    with _engine_lock:
        if not _engine_started:
            lib.StartEngine()
            _engine_started = True

_ensure_engine_started()

class TeaQDM:
    _id_counter = 0
    _lock = threading.Lock()
    _active_bars = {}

    def __init__(self, iterable, desc="Processing", parent=None):
        self.iterable = iterable
        self.desc = desc.encode('utf-8')
        try:
            self.total = len(iterable)
        except TypeError:
            self.total = 0
        
        with TeaQDM._lock:
            TeaQDM._id_counter += 1
            self.id = TeaQDM._id_counter
        
        self.parent = parent
        self._current = 0

    def __iter__(self):
        parent_id = self.parent.id if self.parent and hasattr(self.parent, 'id') else -1
        
        if parent_id >= 0:
            lib.AddBarWithParent(self.id, self.total, self.desc, parent_id)
        else:
            lib.AddBar(self.id, self.total, self.desc)
        
        with TeaQDM._lock:
            TeaQDM._active_bars[self.id] = self
        
        try:
            for item in self.iterable:
                yield item
                self._current += 1
                lib.UpdateBar(self.id, 1)
        finally:
            lib.RemoveBar(self.id)
            with TeaQDM._lock:
                TeaQDM._active_bars.pop(self.id, None)
    
    def update(self, n=1):
        self._current += n
        lib.UpdateBar(self.id, n)
    
    def set_total(self, total):
        self.total = total
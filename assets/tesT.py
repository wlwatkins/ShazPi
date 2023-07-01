#!/usr/bin/python
# -*- coding:utf-8 -*-
import sys
import os

picdir = os.path.join(os.path.dirname(os.path.dirname(os.path.realpath(__file__))), 'pic/2in13')
fontdir = os.path.join(os.path.dirname(os.path.dirname(os.path.realpath(__file__))), 'pic')
libdir = os.path.join(os.path.dirname(os.path.dirname(os.path.realpath(__file__))), 'lib')
if os.path.exists(libdir):
    sys.path.append(libdir)
    
from TP_lib import gt1151
from TP_lib import epd2in13_V3
import time
import logging
from PIL import Image,ImageDraw,ImageFont
import traceback
import threading

logging.basicConfig(level=logging.DEBUG)
flag_t = 1

def pthread_irq() :
    print("pthread running")
    while flag_t == 1 :
        if(gt.digital_read(gt.INT) == 0) :
            GT_Dev.Touch = 1
        else :
            GT_Dev.Touch = 0
        print(GT_Dev.Touch)
    print("thread:exit")



try:
    logging.info("epd2in13_V3 Touch Demo")
    epd = epd2in13_V3.EPD()    
    gt = gt1151.GT1151()
    GT_Dev = gt1151.GT_Development()
    GT_Old = gt1151.GT_Development()
    
    logging.info("init and Clear")
    epd.init(epd.FULL_UPDATE)    
    gt.GT_Init()

    t = threading.Thread(target = pthread_irq)
    t.setDaemon(True)
    t.start()
    # t.join()

    
    while(1):
        # Read the touch input
        gt.GT_Scan(GT_Dev, GT_Old)
        if(GT_Old.X[0] == GT_Dev.X[0] and GT_Old.Y[0] == GT_Dev.Y[0] and GT_Old.S[0] == GT_Dev.S[0]):
            continue
        
        if(GT_Dev.TouchpointFlag):
            
            GT_Dev.TouchpointFlag = 0

except IOError as e:
    logging.info(e)
    
except KeyboardInterrupt:    
    logging.info("ctrl + c:")
    flag_t = 0
    time.sleep(2)
    t.join()
    exit()

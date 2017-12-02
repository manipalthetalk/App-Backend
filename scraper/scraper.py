from selenium import webdriver
from selenium.common.exceptions import TimeoutException
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.common.by import By

from time import sleep
from sys import argv
from bs4 import BeautifulSoup

from pprint import pprint
import json


URL = "http://slcm.manipal.edu/{}"
timeout = 3 

def login(rollno, password):
    """
    Logs the user in and returns the driver.
    Handles wrong credentials, etc. (Returns none in that case)
    """
    driver = webdriver.Chrome()
    driver.get(URL.format('loginform.aspx'))

    user_field = driver.find_element_by_id("txtUserid")
    pass_field = driver.find_element_by_id("txtpassword")

    user_field.send_keys(rollno)
    pass_field.send_keys(password)
    driver.find_element_by_css_selector('#btnLogin').click()

    try:
        elem_present = EC.presence_of_element_located((By.ID, 'lnkBtnHome'))
        WebDriverWait(driver, timeout).until(elem_present)
        return driver

    except TimeoutException:
        return None


def construct(driver):
    """ 
    Main response contructor. Collects responses from 
    independent functions (timetable, attendance, etc) and merges into 
    one final response
    """

    if driver is None:
        return "{ error : 'Request timed out' }"

    driver.get(URL.format('StudentTimeTable.aspx')) ## Get timetable ##
    sleep(1.5) ## Find a better solution
    source = driver.page_source
    ttable = timetable(source)

    driver.get(URL.format('Academics.aspx')) ## Get marks, attendance ##
    try:
        elem_present = EC.presence_of_element_located((By.ID, "ContentPlaceHolder1_rtprAssessmentHeader_pnlInternal_0"))
        WebDriverWait(driver, timeout).until(elem_present)
    except TimeoutException:
        return None

    source = driver.page_source
    driver.close()
    att = attendance(source)
    in_marks = internalmarks(source)
    ex_marks = externalmarks(source)

    response = {"Timetable" : ttable, "Attendance" : att, "Marks" : { "Internal Marks" : in_marks, "External Marks" : ex_marks}}

    return response

def timetable(source):
    """
    Fetches the timetable of the weeek by opening the page.

    Known Bug/Fact : Lets say it is Friday today. The user asks for Mondays
    timetable. The user obviously means the coming Monday, rather than the preceding one. 
    But the timetable data is of the week, hence the timetable for Monday will be of the preceding week.
    """

    soup = BeautifulSoup(source, 'html.parser')
    skeleton = soup.find_all('div', {'class': 'fc-content-skeleton'})
    content_skeleton = skeleton[1]

    week = []

    for td in content_skeleton.find_all('div', {'class': 'fc-content-col'}):
        try:
            day = []
            classes = td.find_all('div', {'class': 'fc-title'})
            timings = td.find_all('div', {'class': 'fc-time'})

            for i in range(len(classes)):
                day.append((timings[i].text, classes[i].text))
        except:
            pass
        week.append(day)

    timetable = {
        "monday": week[0],
        "tuesday": week[1],
        "wednesday": week[2],
        "thursday": week[3],
        "friday": week[4],
        "saturday": week[5],
    }

    return timetable

def attendance(source):
    """
    Subject wise attendance 
    """
    response = {}
    soup = BeautifulSoup(source, 'html.parser')
    table = soup.find('table', {'id' : 'tblAttendancePercentage'})
    subjects = table.find_all('tr')[1:]

    for sub in subjects:
        entries = [i.text for i in sub.find_all('td')]
        response[entries[2]] = { "Total" : entries[4],
                                "Attended" : entries[5],
                                "Missed" : entries[6],
                                "Percentage" : entries[7],
                            }

    return response


def internalmarks(source):
    response = {}
    soup = BeautifulSoup(source, 'html.parser')
    div = soup.find('div', {'id' : 'accordion'})

    sub_names = [sub.text for sub in soup.find_all('a', {'data-parent' : '#accordion'})]
    sub_names = [sub.split('\n')[2] for sub in sub_names]
    sub_names = [' '.join(s.split(' ')[4:]) for s in sub_names]
    sub_names = [s[1:] for s in sub_names]

    #sub_names = list(set(sub_names)) ## Messes with the order of the names
    sub_marks = soup.find_all('div', {'class' : 'panel-collapse collapse'})


    for k, sub in enumerate(sub_marks):
        entries = [i.text for i in sub.find_all('td')]
        resp = {}
        for x in range(0, len(entries) - 2, 3):
            resp[entries[x]] = { "Total" : entries[x+1], "Obtained" : entries[x+2]}

        response[sub_names[k]] = resp

    return response



def externalmarks(source):
    return ' '

def main():
    """ 
    Usage : python scraper.py [regno] [password]

    This scraper will be called by the server.
    """
    regno = argv[1]
    password = argv[2]
    driver = login(regno, password)
    response = construct(driver)
    print(json.dumps(response)) ### ~(^.^)~ pretty printing

main()




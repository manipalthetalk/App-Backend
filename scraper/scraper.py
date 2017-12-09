'''
Authors - Lakshay Kalbhor (kalbhor), Nishant Sahoo (nishantsahoo)
'''

import mechanize
from sys import argv
from bs4 import BeautifulSoup
from pprint import pprint
import json


URL = "http://slcm.manipal.edu/{}"

def login(regno, password):
    """
    Logs the user in and returns the driver.
    Handles wrong credentials, etc. (Returns none in that case)
    """
    driver = mechanize.Browser()
    response = driver.open(URL.format('loginform.aspx'))
    driver.select_form("form1")
    driver.form["txtUserid"] = regno
    driver.form["txtpassword"] = password
    driver.method = "POST"

    response = driver.submit()

    try:
        driver.open(URL.format('Academics.aspx'))
    except Exception:
        return None

    return driver


def construct(driver):
    """ 
    Main response contructor. Collects responses from 
    independent functions (timetable, attendance, etc) and merges into 
    one final response
    """

    if driver is None:
        return "{ error : 'Invalid credentials' }"

    """
    response = driver.open(URL.format('StudentTimeTable.aspx')) ## Get timetable ##
    source = response.read()
    ttable = timetable(source)
    """
    ttable = ""

    response = driver.open(URL.format('Academics.aspx')) ## Get marks, attendance ##
    source = response.read()
    att = attendance(source)
    in_marks = internalmarks(source)

    response = driver.open(URL.format('GradeSheet.aspx')) ## Get marks, attendance ##
    source = response.read()
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

    response = {}
    soup = BeautifulSoup(source, 'html.parser')
    cgpa_span = soup.find('span', {'id' : 'ContentPlaceHolder1_lblCGPA'})
    cgpa = cgpa_span.text

    table = soup.find('table', {'class' : 'table table-bordered'})
    all_tr = table.find_all('tr')
    all_tr = all_tr[1:] # Escape the first tr (Contains table headers like : Title, GPA , Subject name

    for tr in all_tr:
        spans = tr.find_all('span')
        subject_name = spans[1].text
        subject_grade = spans[2].text
        response[subject_name] = subject_grade

    response["total"] = cgpa

    return response

def main():
    """ 
    Usage : python scraper.py [regno] [password]

    This scraper will be called by the server.
    """
    regno = argv[1]
    password = argv[2]
    driver = login(regno, password)
    response = construct(driver)
    print(json.dumps(response))

main()


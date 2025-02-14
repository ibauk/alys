# ALYS - Custom RBLR software

Derived from but independent of ScoreMaster, this software provides comprehensive admin support for the RBLR1000 event run by IBAUK on behalf of the Royal British Legion.

## OVERVIEW
Entrants register for the event in advance using a web capture form hosted on wufoo.com. These details are cleaned using a separate application, Reglist (used for all IBAUK rallies), and loaded into the RBLR database with a status of "Registered" (see below). About a week before the event, all the rider certificates are printed and placed in named envelopes along with "welcome to the IBA" materials. These are brought to Squires en masse. On the Friday of the RBLR weekend, the action physically moves to Squires and entrants are signed in as they arrive. On Saturday morning, signed-in entrants are checked out by team members using phones, iPads, etc. Saturday afternoon onwards, entrants return to Squires and are checked in by team members in the carpark, after which they go through verification and collect their certificates. Any deviations from the preprinted details result in new certificates being printed and posted after the event.

## TEAM MEMBERS
At Squires the event is running by a team from IBAUK plus volunteers from RBL. The RBL people are generally involved in managing things in the carpark, especially the checking-in process when entrants arrive back at Squires. IBA roles include signing in, check out and verification as well as recording known withdrawals throughout the weekend.

## ENTRANT STATUS
In the database each entrant has a status code which is updated following his transition through the system.

- **registered** (aka 'not signed in') - expected but not at Squires yet
- **withdrawn** (before signing in, aka 'DNS') - not expected at Squires
- **signed** in at Squires
- **checked out** - still out riding
- **finisher** - checked in at Squires within 24 hours
- **late finisher** - checked in at Squires after 24 hours
- **DNF** - ride is abandoned, not returning to Squires

## SIGNING-IN (START)
Riders are registered online via Wufoo forms but on arrival at Squires they are signed in by staff who will check and confirm their details, make any amendments necessary and take charge of any cash or cheques handed in and record details of payments made directly to RBL bank accounts or via JustGiving.com. Pre-ordered t-shirts and patches are collected as part of this process.

The list of riders shown at sign-in includes only those with a status of Registered or Withdrawn. The full list is always available via [administration][Edit any entrant]. The list is always shown in order of last name, first name.

## CHECK-OUT (START)
The check-out process involves capturing odo readings and thereby changing a rider's status to 'riding'.

Riders are released initially in cohorts starting at 0500, then at 10 minute intervals until 0530 or some other predetermined time after which the timestamp changes each minute.

The odo capture screen initially shows 05:00 and trips automatically to 05:10 at 05:01, to 05:20 at 05:11 and so on. Odo readings taken before 05:00 will show the rider as leaving at 05:00, not before.

The list of riders shown at check-out includes only those with a status of signed-in.

## CHECK-IN (FINISH)
On return to Squires, riders are checked in by team members in the carpark. Their odo reading is recorded and this action stops the clock on their ride. Their entrant status is updated to either Finisher or Late Finisher and the rider is sent off to prepare his paperwork ready for verification. A facility is provided to stop the real-time clock for two minutes so that, when a group of riders arrives back at once they'll all be given the same finish time.

The list of riders shown at check-in includes only those with a status of checked-out or DNF.

## VERIFICATION (FINISH)
Each rider takes completed paperwork to the ride verifier who confirms the relevant details and rejects or accepts the ride. If a suitable certificate is available, it's given to the rider straightaway otherwise a record is made for reprint and posting.

The list of riders shown at verification includes only those with a status of Finisher or Late Finisher, with CertificateDelivered not "Y".



## SHOW STATS
A button is provided to show the current state of play: numbers are shown for each (non-zero) entrant status and for total funds raised. A separate button shows the status of pre-ordered merchandise.


## MERCHANDISE STATS
A button is provided to show statistics about pre-ordered t-shirts and patches. It is assumed that anyone completing sign-in will also collect pre-ordered merchandise. No facilities are provided to handle any cash transactions or additional orders.

## CONFIGURATION
An administrative facility is provided to control parameters such as start times and whether the application is set for the start (Friday/Saturday AM) or finish (Saturday PM/Sunday).

## IMPORT ENTRANTS
Mechanism to import entrant details from a CSV created by the Reglist entry filtering application. Options include setting the value of *CertificateAvailable* to 'Y' or 'N' and whether to add new entries only or reload the entire list.

## EXPORT RESULTS
Mechanism to export post-event details for inclusion in the main IBA UK Rides database.
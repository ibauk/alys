BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS "config" (
	"DBInitialised"	INTEGER NOT NULL DEFAULT 1,
	"StartTime"	TEXT NOT NULL DEFAULT '05:00',
	"StartCohortMins"	INTEGER NOT NULL DEFAULT 10,
	"ExtraCohorts"	INTEGER NOT NULL DEFAULT 3,
	"RallyStatus"	TEXT NOT NULL DEFAULT 'S'
);
CREATE TABLE IF NOT EXISTS "entrants" (
	"EntrantID"	INTEGER NOT NULL,
	"Bike"	TEXT,
	"BikeReg"	INTEGER,
	"RiderFirst"	TEXT,
	"RiderLast"	TEXT,
	"RiderAddress1"	TEXT,
	"RiderAddress2"	TEXT,
	"RiderTown"	TEXT,
	"RiderCounty"	TEXT,
	"RiderPostcode"	TEXT,
	"RiderCountry"	TEXT DEFAULT 'United Kingdom',
	"RiderIBA"	TEXT,
	"RiderPhone"	TEXT,
	"RiderEmail"	TEXT,
	"PillionFirst"	TEXT,
	"PillionLast"	TEXT,
	"PillionAddress1"	TEXT,
	"PillionAddress2"	TEXT,
	"PillionTown"	TEXT,
	"PillionCounty"	TEXT,
	"PillionPostcode"	TEXT,
	"PillionCountry"	INTEGER DEFAULT 'United Kingdom',
	"PillionIBA"	TEXT,
	"OdoCounts"	TEXT NOT NULL DEFAULT 'M',
	"OdoStart"	INTEGER,
	"OdoFinish"	INTEGER,
	"CorrectedMiles"	TEXT DEFAULT 0,
	"FinishTime"	TEXT,
	"StartTime"	TEXT,
	"EntrantStatus"	INTEGER NOT NULL DEFAULT 0,
	"NoKName"	TEXT,
	"NoKRelation"	TEXT,
	"NoKPhone"	TEXT,
	"EntryDonation"	TEXT,
	"SquiresCheque"	TEXT,
	"SquiresCash"	TEXT,
	"RBLRAccount"	TEXT,
	"JustGivingAmt"	TEXT,
	"JustGivingURL"	TEXT,
	"Route"	TEXT DEFAULT 'A-NCW',
	"PillionEmail"	TEXT,
	"PillionPhone"	TEXT,
	"RiderRBLR"	TEXT,
	"PillionRBLR"	TEXT,
	"Tshirt1"	TEXT,
	"Tshirt2"	TEXT,
	"Patches"	INTEGER DEFAULT 0,
	"FreeCamping"	TEXT NOT NULL DEFAULT 'N',
	"CertificateDelivered"	TEXT NOT NULL DEFAULT 'N',
	"CertificateAvailable"	TEXT NOT NULL DEFAULT 'N',
	PRIMARY KEY("EntrantID")
);
COMMIT;

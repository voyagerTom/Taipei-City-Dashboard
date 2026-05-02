--
-- PostgreSQL database dump
--

-- Dumped from database version 15.10
-- Dumped by pg_dump version 15.12 (Ubuntu 15.12-1.pgdg22.04+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;





--
-- Name: garbage_truck; Type: TABLE; Schema: public; Owner: -
--



CREATE TABLE public.garbage_truck
(
    district character varying(10) COLLATE pg_catalog."default",
    village character varying(20) COLLATE pg_catalog."default",
    squad character varying(20) COLLATE pg_catalog."default",
    serial_number character varying(20) COLLATE pg_catalog."default",
    car_number character varying(15) COLLATE pg_catalog."default",
    route_name character varying(50) COLLATE pg_catalog."default",
    trip_number character varying(10) COLLATE pg_catalog."default",
    arrival_time integer,
    departure_time integer,
    location text COLLATE pg_catalog."default",
    longitude numeric(9,6),
    latitude numeric(9,6)
);

\echo '創建garbage_truck完成'
						 
-- 使用 COPY 指令匯入資料
COPY public.garbage_truck FROM '/docker-entrypoint-initdb.d/garbage_truck.csv' DELIMITER ',' CSV HEADER;
\echo 'garbage_truck 匯入資料完成'

\echo 'garbage_truck資料塞入完成'


---------------------

--
-- Name: 組件1&2_食安檢驗總覽/食安檢驗分佈
--
\echo 
\set currentTBL 'food_inspections'

-- 2. 稽查紀錄主表
CREATE TABLE public.:currentTBL (
   -- 自動遞增的主鍵
    id SERIAL PRIMARY KEY,
    
    -- 原始資料的項次
    seq_number INTEGER,
    
    -- 業者資訊
    vendor_name VARCHAR(255) NOT NULL, -- 業者名稱(市招)
    vendor_address TEXT NOT NULL,      -- 業者地址
    
    -- 產品與稽查內容
    product_name VARCHAR(255),         -- 產品名稱
    inspection_date DATE,              -- 稽查日期 (建議匯入時轉為西元 YYYY-MM-DD)
    inspection_item TEXT,              -- 稽查/檢驗項目
    inspection_result VARCHAR(100),    -- 稽查/檢驗結果
    
    -- 處分資訊
    violation_law TEXT,                -- 違反之食安法條及相關法
    fine_amount INTEGER DEFAULT 0,     -- 裁罰金額
    remarks TEXT,                      -- 備註
    
    -- 紀錄創建時間
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);



\echo 創建 :currentTBL 完成



COPY public.:currentTBL(

    seq_number ,
    vendor_name ,
    vendor_address ,
    product_name ,
    inspection_date ,
    inspection_item ,
    inspection_result ,
    violation_law ,
    fine_amount,
    remarks 

) FROM '/docker-entrypoint-initdb.d/food_inspections.csv' DELIMITER ',' CSV HEADER;


\echo '組件1&2' :currentTBL 匯入資料完成

  

----------------------


--組件1&2_食安檢驗總覽/食安檢驗分佈
-- Name: 抓取新北/台北食安稽核資料(by 行政區)
--
\echo 
\set currentTBL 'food_inspections_dist'

-- 2. 稽查紀錄主表
CREATE TABLE public.:currentTBL(
   -- 自動遞增的主鍵
    id SERIAL PRIMARY KEY,
    
    -- 原始資料的項次
    seq_number INTEGER,
    
    -- 業者資訊
    vendor_name VARCHAR(255) NOT NULL, -- 業者名稱(市招)

  -- 拆出行政區
county  VARCHAR(20) NOT NULL, -- 縣市
dist VARCHAR(20) NOT NULL, -- 行政區

    vendor_address TEXT NOT NULL,      -- 業者地址
    
    -- 產品與稽查內容
    product_name VARCHAR(255),         -- 產品名稱
    inspection_date DATE,              -- 稽查日期 (建議匯入時轉為西元 YYYY-MM-DD)
    inspection_item TEXT,              -- 稽查/檢驗項目
    inspection_result VARCHAR(100),    -- 稽查/檢驗結果
    
    -- 處分資訊
    violation_law TEXT,                -- 違反之食安法條及相關法
    fine_amount INTEGER DEFAULT 0,     -- 裁罰金額
    remarks TEXT,                      -- 備註
    
    -- 紀錄創建時間
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


\echo 創建 :currentTBL 完成



COPY public.:currentTBL(

     -- 原始資料的項次
    seq_number ,
    -- 業者資訊
    vendor_name ,
  -- 拆出行政區
county  ,
dist ,
    vendor_address ,
    -- 產品與稽查內容
    product_name ,
    inspection_date ,
    inspection_item ,
    inspection_result ,

    -- 處分資訊
    violation_law ,
    fine_amount ,
    remarks 

) FROM '/docker-entrypoint-initdb.d/food_inspections_dist.csv' DELIMITER ',' CSV HEADER;


\echo '組件1&2'  :currentTBL 匯入資料完成




  
----------------



--組件3_登革熱疫情概況
-- Name: 雙北近12個月登革熱病媒蚊調查資料
--

\echo 
\set currentTBL 'dengue_invest '

-- 2. 稽查紀錄主表
CREATE TABLE public.:currentTBL (
  -- 系統自動遞增主鍵
    id SERIAL PRIMARY KEY,
    
    -- 日期 (資料源格式為 YYYY/MM/DD，Postgres 可直接辨識)
    record_date DATE NOT NULL,
    
    -- 行政區劃
    county VARCHAR(20) NOT NULL,    -- 縣市 (如：台北市)
    town VARCHAR(20) NOT NULL,      -- 鄉鎮市區 (如：萬華區)
    village VARCHAR(50) NOT NULL,   -- 村里名稱 (如：綠堤里)
    
    -- 村里代碼 (包含連字號，建議用 VARCHAR)
    village_id VARCHAR(50),         -- 村里ID (如：6300700-017)
    
    -- 地理座標 (經緯度建議使用 DECIMAL 確保精確度)
    longitude DECIMAL(12, 9),       -- 經度 (VillageLon)
    latitude DECIMAL(12, 9),        -- 緯度 (VillageLat)
    
    -- 紀錄創建時間
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP

);


\echo 創建 :currentTBL 完成

COPY public.:currentTBL(
	
	record_date, 
    county, 
    town, 
    village, 
    village_id, 
    longitude, 
    latitude


) FROM '/docker-entrypoint-initdb.d/dengue_invest.csv' DELIMITER ',' CSV HEADER;


\echo '組件3' :currentTBL  匯入資料完成




--------------------------

--組件3_登革熱疫情概況
-- Name: 雙北登革熱近12個月每日確定病例統計
--

\echo 
\set currentTBL 'dengue_daily_case'

-- 2. 稽查紀錄主表
CREATE TABLE public.:currentTBL (
 id SERIAL PRIMARY KEY,
    -- 日期欄位
    onset_date DATE,                   -- 發病日
    diagnosis_date DATE,               -- 個案研判日
    report_date DATE,                  -- 通報日
    
    -- 基本人口統計
    gender CHAR(1),                    -- 性別 (M/F)
    age_group VARCHAR(20),             -- 年齡層 (如 20-24)
    
    -- 行政區劃
    city VARCHAR(20),                  -- 居住縣市
    district VARCHAR(20),              -- 居住鄉鎮
    village VARCHAR(20),               -- 居住村里
    village_code VARCHAR(20),          -- 居住村里代碼
    city_code_moi VARCHAR(10),         -- 內政部居住縣市代碼
    district_code_moi VARCHAR(10),     -- 內政部居住鄉鎮代碼
    
    -- 統計區資訊
    min_stat_area VARCHAR(30),         -- 最小統計區
    center_x NUMERIC(12, 9),           -- 最小統計區中心點X (經度)
    center_y NUMERIC(12, 9),           -- 最小統計區中心點Y (緯度)
    stat_area_level_1 VARCHAR(30),     -- 一級統計區
    stat_area_level_2 VARCHAR(30),     -- 二級統計區
    
    -- 感染資訊
    is_imported VARCHAR(5),               -- 是否境外移入 (或用 VARCHAR(2) 存 "是"/"否")
    country_of_infection VARCHAR(50),  -- 感染國家
    case_count INTEGER DEFAULT 1,      -- 確定病例數
    serotype VARCHAR(20)               -- 血清型
);


\echo 創建 :currentTBL 完成



COPY public.:currentTBL(
	onset_date,
	diagnosis_date,
    report_date, 
	gender,
	age_group, 
    city,
	district,
	village,
	min_stat_area,
	center_x, 
    center_y, 
	stat_area_level_1,
	stat_area_level_2,
	is_imported, 
    country_of_infection,
	case_count,
	village_code,
	serotype, 
    city_code_moi,
	district_code_moi
) FROM '/docker-entrypoint-initdb.d/dengue_daily_case.csv' DELIMITER ',' CSV HEADER;


\echo '組件3' :currentTBL  匯入資料完成





-----------



--組件4_登革熱快篩資源分佈
-- Name: 登革熱快篩資源分佈
--

\echo 
\set currentTBL 'hosp_vaccine_info '

-- 2. 稽查紀錄主表
CREATE TABLE public.:currentTBL (
    id           SERIAL PRIMARY KEY,
    -- 地點資訊
    city         VARCHAR(20),          -- 縣市
    hosp_name    VARCHAR(100),         -- 醫療機構名稱
    hosp_address VARCHAR(200),         -- 醫療機構地址
    
    -- 座標資訊 (使用 NUMERIC 以確保精確度)
    latitude     NUMERIC(10, 7),       -- 緯度
    longitude    NUMERIC(10, 7),       -- 經度
    
    -- 聯絡與類型
    hosp_tel     VARCHAR(20),          -- 醫療機構電話
    vaccine_type VARCHAR(50),          -- 疫苗類型 (例如：登革熱疫苗)
    
    -- 時間戳記
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


\echo  創建 :currentTBL 完成



COPY public.:currentTBL(
	city,           -- 縣市
    hosp_name,      -- 醫療機構名稱
    hosp_address,   -- 醫療機構地址

    latitude,       -- 緯度
    longitude,      -- 經度

    hosp_tel,       -- 醫療機構電話
    vaccine_type    -- 疫苗類型（例如：登革熱疫苗）
) FROM '/docker-entrypoint-initdb.d/hosp_vaccine_info.csv' DELIMITER ',' CSV HEADER;


\echo '組件4' :currentTBL  匯入資料完成




-------------


--組件5_麻疹(漢他病毒)概況
-- Name: 雙北麻疹
--

\echo 
\set currentTBL 'measles_case '

-- 2. 稽查紀錄主表
CREATE TABLE public.:currentTBL (
    id                   SERIAL PRIMARY KEY,
    disease_code         VARCHAR(10),        -- 確定病名 (代碼)
    onset_year           SMALLINT,           -- 發病年份
    onset_week           SMALLINT,           -- 發病週別
    city                 VARCHAR(20),        -- 縣市
    district             VARCHAR(20),        -- 鄉鎮
    gender               CHAR(1),            -- 性別
    is_imported          VARCHAR(5) ,         -- 是否為境外移入 (1:是, 0:否)
    age_group            VARCHAR(20),        -- 年齡層
    case_count           INTEGER DEFAULT 0,  -- 確定病例數
    city_code            VARCHAR(10),        -- 縣市別代碼
    district_code        VARCHAR(10),         -- 鄉鎮別代碼
    
    -- 時間戳記
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


\echo 創建 :currentTBL 完成



COPY public.:currentTBL(
	disease_code,
    onset_year,
    onset_week,
    city,
    district,
    gender,
    is_imported,
    age_group,
    case_count,
    city_code,
    district_code
) FROM '/docker-entrypoint-initdb.d/measles_case.csv' DELIMITER ',' CSV HEADER;


\echo '組件5' :currentTBL  匯入資料完成




------------------



--
-- Name: 組件7_急救責任醫院分佈
--
\set currentTBL 'emergency_hospital'

-- 2. 稽查紀錄主表
CREATE TABLE public.:currentTBL (
    id                SERIAL PRIMARY KEY,
    city              VARCHAR(20),          -- 縣市
    facility_name     VARCHAR(100),         -- 醫事機構名稱
    address           VARCHAR(200),         -- 地址
    phone             VARCHAR(20),          -- 電話
    
    -- 座標資訊
    latitude_x        NUMERIC(10, 7),       -- 緯度 (X)
    longitude_y       NUMERIC(10, 7),       -- 經度 (Y)
    
    -- 系統記錄
    created_at        TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


\echo 創建 :currentTBL 完成



COPY public.:currentTBL(
	city,
    facility_name,
    address,
    phone,
    latitude_x,
    longitude_y
) FROM '/docker-entrypoint-initdb.d/emergency_hospital.csv' DELIMITER ',' CSV HEADER;


\echo '組件7' :currentTBL  匯入資料完成



------------------------





--
-- Name: 組件8_健保特約藥局
--
\echo 
\set currentTBL 'parmacy'
-- 2. 稽查紀錄主表
CREATE TABLE public.:currentTBL (
   -- 自動遞增的主鍵
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,        -- 機構名稱
    address TEXT NOT NULL,             -- 地址
    phone VARCHAR(20),                 -- 電話
    longitude NUMERIC(10, 7),          -- x (經度)
    latitude NUMERIC(10, 7),           -- y (緯度)
    
    -- 紀錄創建時間
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);



\echo 創建 :currentTBL 完成



COPY public.:currentTBL(

 name ,
    address ,
    phone ,
    longitude ,
    latitude 

) FROM '/docker-entrypoint-initdb.d/parmacy.csv' DELIMITER ',' CSV HEADER;


\echo '組件8' :currentTBL 匯入資料完成



  

--
--以上為人工添加
--


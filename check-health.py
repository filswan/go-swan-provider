import json
import sys
import os
import subprocess
import toml
import requests
import numbers


# Configuration file check
def check_config():
    sp_config_path = os.path.join(swan_path, "provider/config.toml")
    try:
        os.stat(sp_config_path)
        report.write("  swan-provider config file is ok. \n")
    except FileNotFoundError:
        report.write("  ERROR: swan-provider config file is not exist! \n")

    try:
        market_version = ''
        parsed_toml = toml.load(sp_config_path)
        for key, value in parsed_toml.get('main').items():
            if key == "market_version":
                market_version = value
                break
        if market_version == '1.2':
            os.stat(os.path.join(swan_path, "provider/boost/config.toml"))
            report.write("  boost config file is ok. \n")
    except FileNotFoundError:
        print("boost config file is not exist!")
        report.write("  ERROR: boost config file is not exist! \n")


# Version check.
def check_version():
    print("start check version")
    report.write("  1. " + do_cmd('lotus -v'))
    report.write("  2. " + do_cmd('lotus-miner -v'))
    report.write("  3. " + do_cmd('boostd -v'))
    report.write("  4. " + do_cmd('swan-provider version') + "\n")


def health_aria2():
    print("start check aria2 service")
    try:
        aria2_host = ''
        aria2_port = ''
        aria2_secret = ''
        parsed_toml = toml.load(os.path.join(swan_path, "provider/config.toml"))
        for key, value in parsed_toml.get('aria2').items():
            if key == "aria2_host":
                aria2_host = value
            if key == "aria2_port":
                aria2_port = value
            if key == "aria2_secret":
                aria2_secret = value
        url = 'http://' + aria2_host + ':' + str(aria2_port) + '/jsonrpc'
        data = {
            "jsonrpc": "2.0",
            "id": "qwer",
            "method": "aria2.getVersion",
            "params": ["token:" + aria2_secret]
        }

        headers = {'content-type': 'application/json'}
        r = requests.post(url, json=data, headers=headers, timeout=6)
        if r.status_code != 200:
            report.write(
                "   1. ERROR: Check the aria2 rpc service fail! return data is " + r.content + ", Please check aria2_host、aria2_port、aria2_secret.")
        else:
            report.write("  1. Check the aria2 rpc service success. \n")
    except:
        report.write("  1. ERROR: Check the aria2 rpc service fail! \n")


def health_swan():
    print("start check swan platform connection")
    try:
        api_key = ''
        access_token = ''
        api_url = ''
        parsed_toml = toml.load(os.path.join(swan_path, "provider/config.toml"))
        for key, value in parsed_toml.get('main').items():
            if key == "api_url":
                api_url = value
            if key == "api_key":
                api_key = value
            if key == "access_token":
                access_token = value
        url = api_url + '/user/login_by_apikey'
        data = {
            "apikey": api_key,
            "access_token": access_token
        }
        headers = {'content-type': 'application/json'}
        r = requests.post(url, json=data, headers=headers, timeout=6)
        if r.status_code != 200:
            report.write(
                "  2. ERROR: Check the connectivity with swan platform fail! return data is " + r.content + ", Please check api_key and access_token! \n")
        else:
            report.write("  2. Check the connectivity with swan platform success. \n")
    except:
        report.write("  2. ERROR: Check the connectivity with swan platform fail! \n")


def health_lotus():
    print("start check lotus and lotus-miner")
    client_api_url = ''
    market_api_url = ''
    parsed_toml = toml.load(os.path.join(swan_path, "provider/config.toml"))
    for key, value in parsed_toml.get('lotus').items():
        if key == "client_api_url":
            client_api_url = value
        if key == "market_api_url":
            market_api_url = value
    try:
        headers = {'content-type': 'application/json'}
        data = {
            "jsonrpc": "2.0",
            "method": 'Filecoin.Version',
            "params": [],
            "id": 1
        }
        r_lotus = requests.post(client_api_url, json=data, headers=headers, timeout=6)
        if r_lotus.status_code != 200:
            report.write(
                "  3. ERROR: Check the connectivity with lotus fail! return data is " + r_lotus.text + ", Please check client_api_url and client_api_token! \n")
        else:
            report.write("  3. Check the connectivity with lotus success. \n")
    except:
        report.write("  3. ERROR: Check the connectivity with lotus fail! \n")
    try:
        data = {
            "jsonrpc": "2.0",
            'method': "Filecoin.ActorAddress",
            "params": [],
            "id": 1
        }
        r_miner = requests.post(market_api_url, json=data, headers=headers, timeout=6)
        if r_miner.status_code != 200:
            report.write(
                "  4. ERROR: Check the connectivity with lotus-miner fail! return data is " + r_miner.text + ", Please check market_api_url and market_access_token! \n")
        else:
            report.write("  4. Check the connectivity with lotus-miner success. \n")
    except:
        report.write("  4. ERROR: Check the connectivity with lotus-miner fail! \n")

    try:
        market_version = ''
        for key, value in parsed_toml.get('main').items():
            if key == "market_version":
                market_version = value
                break
        if market_version == '1.2':
            data = {
                "jsonrpc": "2.0",
                'method': "Filecoin.ID",
                "params": [],
                "id": 1
            }
            r_boost = requests.post('http://127.0.0.1:1288/rpc/v0', json=data, headers=headers, timeout=6)
            if r_boost.status_code != 200:
                report.write(
                    "  5. ERROR: Check the connectivity with boost fail! return data is " + r_boost.text + ", Please check boost service process. \n")
            else:
                report.write("  5. Check the connectivity with boost success. \n")
    except:
        report.write("  5. ERROR: Check the connectivity with boost fail! \n")


def check_val():
    print("start check config file")
    try:
        parsed_toml = toml.load(os.path.join(swan_path, "provider/config.toml"))
        market_version = ''
        for key, value in parsed_toml.items():
            if key == "lotus":
                if len(value['client_api_url']) == 0:
                    report.write("  1.  ERROR: lotus.client_api_url is null \n")
                else:
                    report.write("  1.  lotus.client_api_url is ok. \n")
                if len(value['client_api_token']) == 0:
                    report.write("  2.  ERROR: lotus.client_api_token is null \n")
                else:
                    report.write("  2.  lotus.client_api_token is ok. \n")
                if len(value['market_api_url']) == 0:
                    report.write("  3.  ERROR: lotus.market_api_url is null! \n")
                else:
                    report.write("  3.  lotus.market_api_url is ok. \n")
                if len(value['market_access_token']) == 0:
                    report.write("  4.  ERROR: lotus.market_access_token is null! \n")
                else:
                    report.write("  4.  lotus.market_access_token is ok. \n")

            if key == "aria2":
                aria2_candidate_dirs = value['aria2_candidate_dirs']
                if len(value['aria2_download_dir']) == 0:
                    report.write("  5.  ERROR: aria2.aria2_download_dir is null! \n")
                elif not os.path.exists(value['aria2_download_dir']):
                    report.write("  5.  ERROR: aria2.aria2_download_dir not exist, need to be created manually! \n")
                else:
                    report.write("  5.  aria2.aria2_download_dir is ok. \n")
                if not isinstance(aria2_candidate_dirs, (list, str)):
                    report.write(
                        '  6.  ERROR: aria2.aria2_candidate_dirs is null, allow two format: ["/tmp", "/data"] or "/tmp, /data" ! \n')
                elif isinstance(aria2_candidate_dirs, str):
                    results = aria2_candidate_dirs.split(",")
                    for r_dir in results:
                        if not os.path.exists(r_dir.strip()):
                            report.write(
                                "  6.  ERROR: aria2.aria2_candidate_dirs, directory: " + r_dir.strip() + " not exist! \n")
                elif isinstance(aria2_candidate_dirs, list):
                    for r_dir in aria2_candidate_dirs:
                        if not os.path.exists(r_dir.strip()):
                            report.write(
                                "  6.  ERROR: aria2.aria2_candidate_dirs, directory: " + r_dir.strip() + " not exist! \n")
                else:
                    report.write("  6.  aria2.aria2_candidate_dirs is ok. \n")
                if len(value['aria2_host']) == 0:
                    report.write("  7.  ERROR: aria2.aria2_host is null! \n")
                else:
                    report.write("  7.  aria2.aria2_host is ok. \n")
                if not isinstance(value['aria2_port'], numbers.Number):
                    report.write("  8.  ERROR: aria2.aria2_port is null! \n")
                else:
                    report.write("  8.  aria2.aria2_port is ok. \n")
                if len(value['aria2_secret']) == 0:
                    report.write("  9.  ERROR: aria2.aria2_secret is null! \n")
                else:
                    report.write("  9.  aria2.aria2_secret is ok. \n")
                if not isinstance(value['aria2_max_downloading_tasks'], numbers.Number):
                    report.write("  10. ERROR: aria2.aria2_max_downloading_tasks is null! \n")
                else:
                    report.write("  10. aria2.aria2_max_downloading_tasks is ok. \n")

            if key == "main":
                market_version = value['market_version']
                if len(market_version) == 0:
                    report.write("  11. ERROR: main.market_version is null \n")
                elif market_version != '1.1' and market_version != '1.2':
                    report.write("  11. ERROR: main.market_version type can only be 1.1 or 1.2! \n")
                else:
                    report.write("  11. main.market_version is ok. \n")
                if len(value['miner_fid']) == 0:
                    report.write("  12. ERROR: main.miner_fid is null! \n")
                else:
                    report.write("  12. main.miner_fid is ok. \n")
                if not isinstance(value['import_interval'], numbers.Number):
                    report.write("  13. ERROR: main.import_interval is null! \n")
                else:
                    report.write("  13. main.import_interval is ok. \n")
                if not isinstance(value['scan_interval'], numbers.Number):
                    report.write("  14. ERROR: main.scan_interval is null! \n")
                else:
                    report.write("  14. main.scan_interval is ok. \n")
                if not isinstance(value['api_heartbeat_interval'], numbers.Number):
                    report.write("  15. ERROR: main.api_heartbeat_interval is null! \n")
                else:
                    report.write("  15. main.api_heartbeat_interval is ok. \n")
            if key == "bid":
                bid_mode = value['bid_mode']
                expected_sealing_time = value['expected_sealing_time']
                auto_bid_deal_per_day = value['auto_bid_deal_per_day']
                if bid_mode not in [0, 1]:
                    report.write("  16. ERROR: bid.bid_mode can only be 0 or 1! \n")
                else:
                    if bid_mode == 0:
                        report.write(
                            "  16. Only manual bidding orders can be received, not automatic bidding orders. \n")
                    elif bid_mode == 1:
                        report.write(
                            "  16. Only automatic bidding orders can be received, not manual bidding orders. \n")
                if not isinstance(expected_sealing_time, numbers.Number):
                    report.write("  17. ERROR: bid.expected_sealing_time is null! \n")
                elif 1920 > expected_sealing_time > 2880:
                    report.write("  17. ERROR: bid.expected_sealing_time range is [1920~2880]! \n")
                else:
                    report.write("  17. bid.expected_sealing_time is ok. \n")
                if not isinstance(value['start_epoch'], numbers.Number):
                    report.write("  18. ERROR: bid.start_epoch is null! \n")
                else:
                    report.write("  18. bid.start_epoch is ok. \n")
                if not isinstance(auto_bid_deal_per_day, numbers.Number):
                    report.write("  19. ERROR: bid.auto_bid_deal_per_day is null! \n")
                elif auto_bid_deal_per_day < 500:
                    report.write("  19. ERROR: bid.auto_bid_deal_per_day value must be >= 500! \n")
                else:
                    report.write("  19. bid.auto_bid_deal_per_day is ok. \n")
            if market_version == '1.2':
                b_collateral_wallet = ''
                b_publish_wallet = ''
                boost_toml = toml.load(os.path.join(swan_path, "provider/boost/config.toml"))
                for bk, bv in boost_toml.get('Wallets').items():
                    if bk == 'PublishStorageDeals':
                        b_publish_wallet = bv
                    if bk == 'DealCollateral':
                        b_collateral_wallet = bv

                if key == "market":
                    collateral_wallet = value['collateral_wallet']
                    publish_wallet = value['publish_wallet']
                    if len(collateral_wallet) == 0:
                        report.write("  20. ERROR: market.collateral_wallet is null! \n")
                    elif collateral_wallet != b_collateral_wallet:
                        report.write(
                            "  20. market.collateral_wallet is not the same as in the boost configuration file. \n")
                    else:
                        report.write("  20. bid.collateral_wallet is ok. \n")
                    if len(publish_wallet) == 0:
                        report.write("  21. ERROR: market.publish_wallet is null! \n")
                    elif publish_wallet != b_publish_wallet:
                        report.write(
                            "  21. market.publish_wallet is not the same as in the boost configuration file. \n")
                    else:
                        report.write("  21. market.publish_wallet is ok. \n")
                    if len(collateral_wallet) > 0:
                        headers = {'content-type': 'application/json'}
                        data = {
                            "jsonrpc": "2.0",
                            "method": "Filecoin.WalletBalance",
                            "params": [collateral_wallet],
                            "id": 7878
                        }
                        r = requests.post('https://api.node.glif.io/rpc/v0', json=data, headers=headers,
                                          timeout=6)
                        if r.status_code != 200:
                            report.write(
                                "Check the wallet balance fail! return data is " + r.text + ", Please check market.collateral_wallet! \n")
                        else:
                            result = json.loads(r.text)
                            balance = int(result['result']) / (10 ** 18)
                            if balance < 10:
                                report.write("The deal collateral wallet: " + collateral_wallet +
                                             " has a balance of less than 10 FIL, Please charge more than 10 FIL! \n")
                            else:
                                report.write("The deal collateral wallet: " + collateral_wallet + "balance is " +
                                             str(balance) + " FIL. \n")
    except FileNotFoundError:
        print("swan-provider config file is not exist!")


def check_query():
    print("start check query-ask")
    miner_fid = ''
    parsed_toml = toml.load(os.path.join(swan_path, "provider/config.toml"))
    for key, value in parsed_toml.get('main').items():
        if key == "miner_fid":
            miner_fid = value
            break
    headers = {'content-type': 'application/json'}
    r = requests.get('https://api.filswan.com/tools/check_connectivity?storage_provider_id=' + miner_fid,
                     headers=headers, timeout=6)
    if r.status_code != 200:
        report.write("  ERROR: Check miner query-ask failed! return data is " + r.text + " \n")
    else:
        parsed_toml = toml.load(os.path.join(swan_path, "provider/config.toml"))
        market_version = ''
        for key, value in parsed_toml.get('main').items():
            if key == "market_version":
                market_version = value
                break
        result = json.loads(r.text)
        if result['status'] == "fail":
            if market_version == '1.1':
                report.write("  ERROR: Please use the command 'lotus-miner net listen' to query multi-address! \n")
            if market_version == '1.2':
                report.write(
                    "  ERROR: Please use the command 'boostd --boost-repo=$SWAN_PATH/provider/boost net listen' to query multi-address! \n")
        elif result['status'] == "success":
            if result['data']['price_per_GiB'] != "0" or result['data']['verified_price_per_GiB'] != "0":
                if market_version == '1.1':
                    report.write("  ERROR: Please use the command 'lotus-miner storage-deals set-ask --price 0 "
                                 "--verified-price 0 --min-piece-size 56KiB --max-piece-size 64GB' to set the price! \n")
                if market_version == '1.2':
                    report.write("  ERROR: Please use the command 'export SWAN_PATH=$SWAN_PATH && swan-provider "
                                 "set-ask --price=0 --verified-price=0 --min-piece-size=256 "
                                 "--max-piece-size=34359738368' to set the price! \n")
            else:
                report.write("  Check miner query-ask is ok. \n")


def do_cmd(cmd_str):
    try:
        p = subprocess.Popen(cmd_str, shell=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
        (stdout, stderr) = p.communicate()
        if stdout:
            return stdout.decode('utf-8')
        if stderr:
            return "ERROR: "+stderr.decode('utf-8')
    except subprocess.TimeoutExpired:
        p.kill()


def do_cmd_out(cmd_str):
    try:
        p = subprocess.Popen(cmd_str, shell=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
        (stdout, stderr) = p.communicate()
        if stdout:
            lines = stdout.decode('utf-8').split('\n')
            for line in lines:
                new_line = line.strip()
                if new_line.startswith("Active:") is True:
                    split_data = new_line.split(":")
                    if split_data[1].find("running") != -1:
                        return "Check the aria2c service success."
                    else:
                        return "Check the aria2c service fail! Please start aria2c service."
                    break
        if stderr:
            return stderr.decode('utf-8')
    except subprocess.TimeoutExpired:
        p.kill()


if __name__ == '__main__':
    if len(sys.argv) < 2:
        print("The command line must specify a parameter to be used as swan_path！")
    else:
        swan_path = sys.argv[1]
        report = open("report.txt", "w")
        report.write("Version: \n")
        check_version()
        report.write("Config: \n")
        check_config()
        check_val()
        report.write("\n")
        report.write("Service: \n")
        health_aria2()
        health_swan()
        health_lotus()
        report.write("\n")
        report.write("Query-ask: \n")
        check_query()
        report.close()

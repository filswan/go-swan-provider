import json
import os
import subprocess
import toml
import requests


# Configuration file check
def check_config():
    sp_config_path = os.path.join(swan_path, "provider/config.toml")
    try:
        os.stat(sp_config_path)
    except FileNotFoundError:
        print("swan-provider config file is not exist!")
    try:
        market_version = ''
        parsed_toml = toml.load(sp_config_path)
        for key, value in parsed_toml.get('main').items():
            if key == "market_version":
                market_version = value
                break
        if market_version == '1.2':
            os.stat(os.path.join(swan_path, "provider/boost/config.toml"))
    except FileNotFoundError:
        print("boost config file is not exist!")


# Version check.
def check_version():
    do_cmd('lotus -v')
    do_cmd('lotus-miner -v')
    do_cmd('boostd -v')
    do_cmd('swan-provider version')
    do_cmd_out('systemctl status aria2c')


def health_swan():
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
    url = api_url+'/user/login_by_apikey'
    data = {
        "apikey": api_key,
        "access_token": access_token
    }
    headers = {'content-type': 'application/json'}
    r = requests.post(url, json=data, headers=headers)
    if r.status_code != 200:
        print("Check the connectivity with swan platform fail! Please check api_key and access_token.")
    else:
        print("Check the connectivity with swan platform success.")


def health_aria2():
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
    url = 'http://'+aria2_host+':'+str(aria2_port)+'/jsonrpc'
    data = {
        "jsonrpc": "2.0",
        "id": "qwer",
        "method": "aria2.getVersion",
        "params": ["token:"+aria2_secret]
    }

    headers = {'content-type': 'application/json'}
    r = requests.post(url, json=data, headers=headers)
    if r.status_code != 200:
        print("Check the aria2 rpc service fail! Please check aria2_host、aria2_port、aria2_secret.")
        print(r.content)
        print(r.request)
    else:
        print("Check the aria2 rpc service success.")


def health_lotus():
    client_api_url = ''
    market_api_url = ''
    parsed_toml = toml.load(os.path.join(swan_path, "provider/config.toml"))
    for key, value in parsed_toml.get('lotus').items():
        if key == "client_api_url":
            client_api_url = value
        if key == "market_api_url":
            market_api_url = value

    headers = {'content-type': 'application/json'}
    data = {
        "jsonrpc": "2.0",
        "method": 'Filecoin.Version',
        "params": [],
        "id": 1
    }
    r_lotus = requests.post(client_api_url, json=data, headers=headers)
    if r_lotus.status_code != 200:
        print("Check the connectivity with lotus fail! Please check client_api_url and client_api_token.")
        print(r_lotus.text)
    else:
        print("Check the connectivity with lotus success.")

    data = {
        "jsonrpc": "2.0",
        'method': "Filecoin.ActorAddress",
        "params": [],
        "id": 1
    }
    r_miner = requests.post(market_api_url, json=data, headers=headers)
    if r_miner.status_code != 200:
        print("Check the connectivity with lotus-miner fail! Please check market_api_url and market_access_token.")
        print(r_miner.text)
    else:
        print("Check the connectivity with lotus-miner success.")

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
        r_boost = requests.post('http://127.0.0.1:1288/rpc/v0', json=data, headers=headers)
        if r_boost.status_code != 200:
            print("Check the connectivity with boost fail! Please check boost service process.")
            print(r_boost.text)
        else:
            print("Check the connectivity with boost success.")


def check_val():
    parsed_toml = toml.load(os.path.join(swan_path, "provider/config.toml"))
    market_version = ''
    for key, value in parsed_toml.items():
        if key == "lotus":
            if print_log(value['client_api_token']) is False:
                print("lotus.client_api_token is null")

            if print_log(value['market_access_token']) is False:
                print("lotus.market_access_token is null")
        if key == "aria2":
            aria2_download_dir = value['aria2_download_dir']
            aria2_candidate_dirs = value['aria2_candidate_dirs']
            aria2_host = value['aria2_host']
            aria2_port = value['aria2_port']
            aria2_secret = value['aria2_secret']
            aria2_max_downloading_tasks = value['aria2_max_downloading_tasks']
            if print_log(aria2_download_dir) is False:
                print("aria2.aria2_download_dir is null")
            if print_log(aria2_candidate_dirs) is False:
                print("aria2.aria2_candidate_dirs is null")
            if print_log(aria2_host) is False:
                print("aria2.aria2_host is null")
            if print_log(aria2_port) is False:
                print("aria2.aria2_port is null")
            if print_log(aria2_secret) is False:
                print("aria2.aria2_secret is null")
            if print_log(aria2_max_downloading_tasks) is False:
                print("aria2.aria2_max_downloading_tasks is null")
        if key == "main":
            market_version = value['market_version']
            miner_fid = value['miner_fid']
            import_interval = value['import_interval']
            scan_interval = value['scan_interval']
            api_heartbeat_interval = value['api_heartbeat_interval']
            if print_log(market_version) is False:
                print("main.market_version is null")
            elif market_version != '1.1' and market_version != '1.2':
                print("main.market_version type can only be 1.1 or 1.2")
            if print_log(miner_fid) is False:
                print("main.miner_fid is null")
            if print_log(import_interval) is False:
                print("main.import_interval is null")
            if print_log(scan_interval) is False:
                print("main.scan_interval is null")
            if print_log(api_heartbeat_interval) is False:
                print("main.api_heartbeat_interval is null")
        if key == "bid":
            bid_mode = value['bid_mode']
            expected_sealing_time = value['expected_sealing_time']
            start_epoch = value['start_epoch']
            auto_bid_deal_per_day = value['auto_bid_deal_per_day']
            if print_log(bid_mode) is False:
                print("bid.bid_mode is null")
            else:
                if bid_mode == 0:
                    print('Only manual bidding orders can be received, not automatic bidding orders.')
                elif bid_mode == 1:
                    print("Only automatic bidding orders can be received, not manual bidding orders.")
                if bid_mode != 1 and bid_mode != 2:
                    print("bid_mode=", bid_mode, "value unknown!")
            if print_log(expected_sealing_time) is False:
                print("bid.expected_sealing_time is null")
            elif 1920 > expected_sealing_time > 2880:
                print("bid.expected_sealing_time range is [1920~2880]")
            if print_log(start_epoch) is False:
                print("bid.start_epoch is null")
            if print_log(auto_bid_deal_per_day) is False:
                print("bid.auto_bid_deal_per_day is null")
            elif auto_bid_deal_per_day < 500:
                print("bid.auto_bid_deal_per_day value must be >= 500")
        if market_version == '1.2':
            if key == "market":
                collateral_wallet = value['collateral_wallet']
                publish_wallet = value['publish_wallet']
                if print_log(collateral_wallet) is False:
                    print("market.collateral_wallet is null")
                if print_log(publish_wallet) is False:
                    print("market.publish_wallet is null")

                if print_log(collateral_wallet) & print_log(publish_wallet):
                    headers = {'content-type': 'application/json'}
                    data = {
                        "jsonrpc": "2.0",
                        "method": "Filecoin.WalletBalance",
                        "params": ["\""+collateral_wallet+"\""],
                        "id": 7878
                    }
                    r = requests.post('https://api.calibration.node.glif.io/rpc/v0', json=data, headers=headers)
                    if r.status_code != 200:
                        print("Check the wallet balance fail! Please check market.collateral_wallet.")
                    else:
                        result = json.loads(r.text)
                        balance = ("%.2f" % int(result['result'])/10^18)
                        if balance < 10:
                            print("The deal collateral wallet",collateral_wallet,
                                  'has a balance of less than 10FIL, Please charge more than 10FIL')
                        else:
                            print("The deal collateral wallet", collateral_wallet, 'balance is',balance,'FIL, ok.')


def check_query():
    miner_fid = ''
    parsed_toml = toml.load(os.path.join(swan_path, "provider/config.toml"))
    for key, value in parsed_toml.get('main').items():
        if key == "miner_fid":
            miner_fid = value
            break
    headers = {'content-type': 'application/json'}
    r = requests.get('https://calibration-api.filswan.com/tools/check_connectivity?storage_provider_id='+miner_fid,
                     headers=headers)
    if r.status_code != 200:
        print("check miner query-ask failed, status is", r.status_code)
    else:
        parsed_toml = toml.load(os.path.join(swan_path, "provider/config.toml"))
        market_version = ''
        for key, value in parsed_toml.get('main').items():
            if key == "market_version":
                market_version = value
                break
        result = json.loads(r.text)
        if result['status'] == "fail":
            multi_address = ''
            if multi_address is None:
                print("Please set Miner multi-address!")
            else:
                if market_version == '1.1':
                    print("lotus-miner net listen")
                if market_version == '1.2':
                    print("boostd --boost-repo=$SWAN_PATH/provider/boost net listen")
        elif result['status'] == "success":
            if result['data']['price_per_GiB'] != "0" or result['data']['verified_price_per_GiB'] != "0":
                if market_version == '1.1':
                    do_cmd("lotus-miner storage-deals set-ask --price 0 --verified-price 0 --min-piece-size 56KiB "
                           "--max-piece-size 64GB")
                if market_version == '1.2':
                    my_env = os.environ.copy()
                    my_env["SWAN_PATH"] = swan_path
                    do_cmd('swan-provider set-ask --price=0 --verified-price=0 --min-piece-size=256 '
                           '--max-piece-size=34359738368')


def print_log(field_data):
    if field_data is None:
        # print('Require', field_data, 'field, Please check ',field_data)
        return False
    return True


def do_cmd(cmd_str):
    try:
        p = subprocess.Popen(cmd_str, shell=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
        (stdout, stderr) = p.communicate()
        if stdout:
            print(stdout.decode('utf-8'))
        if stderr:
            print(stderr.decode('utf-8'))
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
                        print("Check the aria2c service success.")
                    else:
                        print("Check the aria2c service fail! Please start aria2c service.")
                    break
        if stderr:
            print(stderr.decode('utf-8'))
    except subprocess.TimeoutExpired:
        p.kill()


if __name__ == '__main__':
    swan_path = "/data/.swan"
    check_version()
    check_config()
    check_val()
    health_aria2()
    health_swan()
    health_lotus()
    check_query()

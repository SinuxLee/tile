#!/usr/bin/env bash
ulimit -c unlimited
export GOTRACEBACK=crash

workspace=$(cd "$(dirname $0)/" || exit 1; pwd)
cd ${workspace} || exit 1

app_name=tms
pid_file=app.pid

function check_pid() {
    if [[ -f ${pid_file} ]];then
        pid=`cat ${pid_file}`
        if [[ -n ${pid} ]]; then
            running=`ps -p ${pid}|grep -v "PID TTY" |wc -l`
            return ${running}
        fi
    fi
    return 0
}

function start() {
    check_pid
    running=$?
    if [[ ${running} -gt 0 ]];then
        echo -n "$app_name now is running already, pid="
        return 1
    fi

    chmod u+x ./${app_name}
    nohup ./${app_name}  > /dev/null 2>&1 &
    sleep 1
    running=`ps -p $! | grep -v "PID TTY" | wc -l`
    if [[ ${running} -gt 0 ]];then
        echo $! > ${pid_file}
        echo "$app_name started..., pid=$!"
    else
        echo "$app_name failed to start."
        return 1
    fi
}

function stop() {
    pid=`cat ${pid_file}`
    kill -15 ${pid}
    rm -f ${pid_file}
    echo "$app_name stoped..."
}

function restart() {
    stop
    sleep 1
    start
}

function status() {
    check_pid
    running=$?
    if [[ ${running} -gt 0 ]];then
        echo started
    else
        echo stoped
    fi
}

function update(){
    file=${app_name}.tar.gz
    tar -xf $file
    backup=${app_name}_$(date +%m_%d).tar.gz
    mv -f $file backup/$backup
    restart
}

function help_me() {
    echo "$0 start|stop|restart|status"
}

if [[ "$1" == "" ]]; then
    help_me
elif [[ "$1" == "stop" ]];then
    stop
elif [[ "$1" == "start" ]];then
    start
elif [[ "$1" == "restart" ]];then
    restart
elif [[ "$1" == "status" ]];then
    status
elif [[ "$1" == "update" ]];then
    update
else
    help_me
fi

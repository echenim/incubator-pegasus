/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

#include <nlohmann/json.hpp>
#include <nlohmann/json_fwd.hpp>
#include <pegasus/git_commit.h>
#include <pegasus/version.h>
#include <unistd.h>
#include <cstdio>
#include <map>
#include <memory>
#include <string>
#include <vector>

#include "backup_types.h"
#include "common/replication_common.h"
#include "compaction_operation.h"
#include "info_collector_app.h"
#include "meta/meta_service_app.h"
#include "pegasus_server_impl.h"
#include "pegasus_service_app.h"
#include "runtime/app_model.h"
#include "runtime/service_app.h"
#include "utils/command_manager.h"
#include "utils/fmt_logging.h"
#include "utils/process_utils.h"
#include "utils/strings.h"
#include "utils/time_utils.h"
#include "utils/utils.h"

#define STR_I(var) #var
#define STR(var) STR_I(var)
#ifndef DSN_BUILD_TYPE
#define PEGASUS_BUILD_TYPE ""
#else
#define PEGASUS_BUILD_TYPE STR(DSN_BUILD_TYPE)
#endif

static char const rcsid[] =
    "$Version: Pegasus Server " PEGASUS_VERSION " (" PEGASUS_GIT_COMMIT ")"
#if defined(DSN_BUILD_TYPE)
    " " STR(DSN_BUILD_TYPE)
#endif
        ", built by gcc " STR(__GNUC__) "." STR(__GNUC_MINOR__) "." STR(__GNUC_PATCHLEVEL__)
#if defined(DSN_BUILD_HOSTNAME)
            ", built on " STR(DSN_BUILD_HOSTNAME)
#endif
                ", built at " __DATE__ " " __TIME__ " $";

const char *pegasus_server_rcsid() { return rcsid; }

using namespace dsn;
using namespace dsn::replication;

void dsn_app_registration_pegasus()
{
    dsn::service::meta_service_app::register_components();
    service_app::register_factory<pegasus::server::pegasus_meta_service_app>("meta");
    service_app::register_factory<pegasus::server::pegasus_replication_service_app>(
        dsn::replication::replication_options::kReplicaAppType.c_str());
    service_app::register_factory<pegasus::server::info_collector_app>("collector");
    pegasus::server::pegasus_server_impl::register_service();
    pegasus::server::register_compaction_operations();
}

int main(int argc, char **argv)
{
    for (int i = 1; i < argc; ++i) {
        if (utils::equals(argv[i], "-v") || utils::equals(argv[i], "-version") ||
            utils::equals(argv[i], "--version")) {
            printf("Pegasus Server %s (%s) %s\n",
                   PEGASUS_VERSION,
                   PEGASUS_GIT_COMMIT,
                   PEGASUS_BUILD_TYPE);
            dsn_exit(0);
        }
    }
    LOG_INFO("pegasus server starting, pid({}), version({})", getpid(), pegasus_server_rcsid());
    dsn_app_registration_pegasus();

    std::unique_ptr<command_deregister> server_info_cmd =
        dsn::command_manager::instance().register_single_command(
            "server-info",
            "Query server information",
            "",
            [](const std::vector<std::string> &args) {
                nlohmann::json info;
                info["version"] = PEGASUS_VERSION;
                info["build_type"] = PEGASUS_BUILD_TYPE;
                info["git_SHA"] = PEGASUS_GIT_COMMIT;
                info["start_time"] =
                    ::dsn::utils::time_s_to_date_time(dsn::utils::process_start_millis() / 1000);
                return info.dump();
            });

    dsn_run(argc, argv, true);

    return 0;
}

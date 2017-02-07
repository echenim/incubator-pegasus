#!/bin/sh

if [ -z "${REPORT_DIR}" ]; then
    REPORT_DIR="."
fi

./clear.sh
output_xml="${REPORT_DIR}/dsn.meta.test.1.xml"
GTEST_OUTPUT="xml:${output_xml}" GTEST_FILTER="meta.state_sync:meta.update_configuration:meta.balancer_validator:meta.simple_lb_balanced_cure:meta.simple_lb_cure_test:meta.json_compacity" ./dsn.meta.test

./clear.sh
output_xml="${REPORT_DIR}/dsn.meta.test.2.xml"
GTEST_OUTPUT="xml:${output_xml}" GTEST_FILTER="meta.data_definition" ./dsn.meta.test config-ddl-test.ini

./clear.sh
output_xml="${REPORT_DIR}/dsn.meta.test.3.xml"
GTEST_OUTPUT="xml:${output_xml}" GTEST_FILTER="meta.apply_balancer" ./dsn.meta.test


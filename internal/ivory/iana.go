/*
 Copyright 2021 - 2023 Highgo Solutions, Inc.
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

 http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package ivory

// The protocol used by IvorySQL is registered with the Internet Assigned
// Numbers Authority (IANA).
// - https://www.iana.org/assignments/service-names-port-numbers
const (
	// IANAPortNumber is the port assigned to IvorySQL at the IANA.
	IANAPortNumber = 5432

	// IANAServiceName is the name of the IvorySQL protocol at the IANA.
	IANAServiceName = "ivorysql"
)

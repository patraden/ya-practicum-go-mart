from diagrams import Cluster, Diagram, Edge
from diagrams.onprem.queue import Kafka
from diagrams.onprem.database import PostgreSQL
from diagrams.programming.language import Go
from diagrams.generic.database import SQL
from diagrams.generic.storage import Storage

graph_attr = {
    "splines": "line",
    "beautify": "true",
    "fontsize": "16",
    # "bgcolor": "transparent",
}

kafka_cluster_attr = {
    "bgcolor": "lightblue",
    "style": "rounded",
    "pencolor": "grey",
    "fontsize": "16",
    "beautify": "true",
}

database_cluster_attr = {
    "bgcolor": "lightblue",
    "style": "rounded",
    "pencolor": "grey",
    "fontsize": "16",
    "beautify": "true"
}

accruals_cluster_attr = {
    "bgcolor": "lightblue",
    "style": "rounded",
    "pencolor": "grey",
    "fontsize": "16",
    "beautify": "true"
}

gophermart_cluster_attr = {
    "bgcolor": "#EBF3E7",
    "pencolor": "grey",
    "style": "rounded",
    "fontsize": "16",
    "beautify": "true"
}


def main():
    with Diagram("Gophermart Architecture", show=False, direction="TB", graph_attr=graph_attr):
        
        with Cluster("Database", graph_attr=database_cluster_attr):
            pg = PostgreSQL()
            orders = SQL("Orders")
            users = SQL("Users")
            balances = SQL("Balances")
            withdraws = SQL("Withdraws")
        
        with Cluster("Kafka Queue", graph_attr=kafka_cluster_attr):
            kafka_pending = Kafka("Pending")
            kafka_processed = Kafka("Processed")
            kafka_dql = Kafka("DLQ")
        
        with Cluster("Accruals Service", graph_attr=accruals_cluster_attr):
            haccruals = Go("GET /api/orders/{number}")
        
        with Cluster("Storage", graph_attr=kafka_cluster_attr):
            fs = Storage("file storage")
        
        with Cluster("Gophermart handlers", graph_attr=gophermart_cluster_attr):
            hregister = Go("POST Register")
            hlogin = Go("POST Login")
            hporders = Go("POST Orders")
            hgorders = Go("GET Orders")
            hbalance = Go("GET Balance")
            hwithdraw = Go("POST Withdraw")
            hwithdraws = Go("GET Withdraws")
            orders_buffer = Go("Orders buffer")
            accruals_processor = Go("Orders updater")
            
            #  edges in Gophermart
            hporders >> Edge(style="dashed") >> orders_buffer
            hlogin >> Edge(color="transparent") >> accruals_processor
        
        with Cluster("Gophermart integrations", graph_attr=gophermart_cluster_attr):
            ad_accruals = Go("Accruals Adapter")
            dlq_sanitizer = Go("DLQ Sanitizer")
            
        # Edges for Kafka and fs
        hporders >> Edge(style="dashed") >> kafka_pending
        accruals_processor << Edge(style="dashed") << kafka_processed
        orders_buffer >> Edge(style="dashed") >> kafka_pending
        orders_buffer >> Edge(style="dashed") >> fs
        dlq_sanitizer >> Edge(style="dashed") >> kafka_pending
        dlq_sanitizer << Edge(style="dashed") << kafka_dql
        
        # Edges for database
        orders << Edge(style="dashed") << [hporders, hgorders, accruals_processor]
        users << Edge(style="dashed") << [hregister, hlogin]
        balances << Edge(style="dashed") << hbalance
        withdraws << Edge(style="dashed") << hwithdraws
        balances << Edge(style="dashed") << hwithdraw
        withdraws << Edge(style="dashed") << hwithdraw

            
        # Edges for Accruals
        kafka_pending >> Edge(style="dashed") >> ad_accruals
        kafka_processed << Edge(style="dashed") << ad_accruals
        kafka_dql << Edge(style="dashed") << ad_accruals
        haccruals << Edge(style="dashed") << ad_accruals


if __name__ == "__main__":
    main()
